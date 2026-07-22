package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/pkg/logger"
	"github.com/acnoway/litellm-go-gateway/internal/pkg/metrics"
	"github.com/google/uuid"
)

// ChatService 是聊天用例的编排层。负责模型路由、重试、fallback 和用量统计。
type ChatService struct {
	providerManager interface {
		Chat(context.Context, biz.ChatRequest) (biz.ChatResponse, error)
		ChatStream(context.Context, biz.ChatRequest) (biz.ChatStream, error)
		GetProvidersForModel(string) []biz.Provider
		Name() string
	}
	retryConfig config.RetryConfig
	usageRepo   biz.UsageRepo
}

// NewChatService 通过接口注入 provider manager，便于替换实现和单元测试。
func NewChatService(providerManager interface {
	Chat(context.Context, biz.ChatRequest) (biz.ChatResponse, error)
	ChatStream(context.Context, biz.ChatRequest) (biz.ChatStream, error)
	GetProvidersForModel(string) []biz.Provider
	Name() string
}, retryConfig config.RetryConfig, usageRepo biz.UsageRepo) *ChatService {
	return &ChatService{
		providerManager: providerManager,
		retryConfig:     retryConfig,
		usageRepo:       usageRepo,
	}
}

// Complete 执行非流式聊天调用。ctx 来自 HTTP 请求，客户端取消时会传递到上游请求。
// 支持自动 fallback：如果主 provider 失败，会自动尝试备用 providers。
func (s *ChatService) Complete(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	log := logger.FromContext(ctx)
	requestID := uuid.New().String()
	startTime := time.Now()

	providers := s.providerManager.GetProvidersForModel(request.Model)
	if len(providers) == 0 {
		return biz.ChatResponse{}, fmt.Errorf("no provider available for model %s", request.Model)
	}

	log.Info("starting chat completion",
		"model", request.Model,
		"stream", false,
		"primary_provider", providers[0].Name(),
		"fallback_count", len(providers)-1,
		"request_id", requestID,
	)

	var lastErr error
	var resp biz.ChatResponse

	// 依次尝试所有可用的 providers
	for i, provider := range providers {
		providerName := provider.Name()
		providerStartTime := time.Now()

		if i > 0 {
			log.Info("trying fallback provider",
				"provider", providerName,
				"attempt", i+1,
				"total_providers", len(providers),
				"request_id", requestID,
			)
		}

		resp, lastErr = s.withRetry(ctx, func() (biz.ChatResponse, error) {
			return provider.Chat(ctx, request)
		})

		providerDuration := time.Since(providerStartTime)

		if lastErr == nil {
			// 成功，记录使用日志、metrics 并返回
			duration := time.Since(startTime).Milliseconds()
			log.Info("chat completion succeeded",
				"provider", providerName,
				"request_id", requestID,
			)

			// 记录 provider 调用成功的 metrics
			metrics.RecordProviderCall(providerName, request.Model, "success", providerDuration)

			s.recordUsage(ctx, requestID, request.Model, resp.Body, true, "", duration, providerName)
			return resp, nil
		}

		// 失败，记录错误和 metrics
		log.Warn("provider failed",
			"provider", providerName,
			"error", lastErr,
			"request_id", requestID,
		)

		// 记录 provider 调用失败的 metrics
		metrics.RecordProviderCall(providerName, request.Model, "error", providerDuration)
	}

	// 所有 providers 都失败了
	duration := time.Since(startTime).Milliseconds()
	errorCode := ""
	var providerErr *biz.ProviderError
	if errors.As(lastErr, &providerErr) {
		errorCode = providerErr.Code
	}

	log.Error("all providers failed",
		"error", lastErr,
		"request_id", requestID,
		"providers_tried", len(providers),
	)

	// 记录失败的最后一个 provider 名称
	lastProviderName := ""
	if len(providers) > 0 {
		lastProviderName = providers[len(providers)-1].Name()
	}

	s.recordUsage(ctx, requestID, request.Model, nil, false, errorCode, duration, lastProviderName)
	return biz.ChatResponse{}, lastErr
}

// CompleteStream 执行流式聊天调用，并把仍打开的上游流交给 Handler 转发。
// 流式请求不重试，因为响应体已经开始传输，中途重试会导致客户端收到重复或错误的数据。
// 但支持 fallback：如果主 provider 在建立流之前失败，会尝试备用 providers。
func (s *ChatService) CompleteStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	log := logger.FromContext(ctx)

	providers := s.providerManager.GetProvidersForModel(request.Model)
	if len(providers) == 0 {
		return biz.ChatStream{}, fmt.Errorf("no provider available for model %s", request.Model)
	}

	log.Info("starting chat stream",
		"model", request.Model,
		"stream", true,
		"primary_provider", providers[0].Name(),
		"fallback_count", len(providers)-1,
	)

	var lastErr error

	// 依次尝试所有可用的 providers
	for i, provider := range providers {
		providerName := provider.Name()
		if i > 0 {
			log.Info("trying fallback provider for stream",
				"provider", providerName,
				"attempt", i+1,
				"total_providers", len(providers),
			)
		}

		stream, err := provider.ChatStream(ctx, request)
		if err == nil {
			log.Info("chat stream started",
				"provider", providerName,
			)
			return stream, nil
		}

		lastErr = err
		log.Warn("provider stream failed",
			"provider", providerName,
			"error", err,
		)
	}

	// 所有 providers 都失败了
	log.Error("all providers failed for stream",
		"error", lastErr,
		"providers_tried", len(providers),
	)
	return biz.ChatStream{}, lastErr
}

// withRetry 实现指数退避重试逻辑。只对网络错误（连接失败、超时、DNS 解析失败）重试，
// 对于 4xx 客户端错误或 5xx 服务端错误不重试，因为立即重试不太可能成功。
func (s *ChatService) withRetry(ctx context.Context, fn func() (biz.ChatResponse, error)) (biz.ChatResponse, error) {
	log := logger.FromContext(ctx)
	var lastErr error
	delay := s.retryConfig.InitialDelay

	for attempt := 1; attempt <= s.retryConfig.MaxAttempts; attempt++ {
		resp, err := fn()
		if err == nil {
			if attempt > 1 {
				log.Info("retry succeeded",
					"attempt", attempt,
					"total_attempts", s.retryConfig.MaxAttempts,
				)
			}
			return resp, nil
		}

		// 不重试的错误类型
		var providerErr *biz.ProviderError
		if errors.As(err, &providerErr) {
			// Provider 返回的 HTTP 错误不重试（4xx/5xx）
			return biz.ChatResponse{}, err
		}

		// 只重试网络相关错误
		if !isRetryableError(err) {
			return biz.ChatResponse{}, err
		}

		lastErr = err

		// 最后一次尝试失败后不再等待
		if attempt == s.retryConfig.MaxAttempts {
			log.Error("all retry attempts exhausted",
				"attempt", attempt,
				"error", err,
			)
			break
		}

		log.Warn("retryable error occurred, will retry",
			"attempt", attempt,
			"max_attempts", s.retryConfig.MaxAttempts,
			"error", err,
			"next_delay_ms", delay.Milliseconds(),
		)

		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return biz.ChatResponse{}, ctx.Err()
		case <-time.After(delay):
		}

		// 指数退避，每次翻倍，但不超过最大延迟
		delay *= 2
		if delay > s.retryConfig.MaxDelay {
			delay = s.retryConfig.MaxDelay
		}
	}

	return biz.ChatResponse{}, lastErr
}

// isRetryableError 判断错误是否值得重试。
// 网络层错误（连接失败、超时、DNS 失败）应该重试，HTTP 错误码不应重试。
func isRetryableError(err error) bool {
	// 网络操作错误
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// 系统调用错误（连接被拒绝、网络不可达等）
	var syscallErr *net.OpError
	if errors.As(err, &syscallErr) {
		if errors.Is(syscallErr.Err, syscall.ECONNREFUSED) ||
			errors.Is(syscallErr.Err, syscall.ECONNRESET) ||
			errors.Is(syscallErr.Err, syscall.ENETUNREACH) {
			return true
		}
	}

	// HTTP 客户端超时错误
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

// recordUsage 从响应体中提取 token 使用信息并记录到数据库和 metrics。
// 响应体应该是 OpenAI 格式的 JSON，包含 usage 字段。
func (s *ChatService) recordUsage(ctx context.Context, requestID string, model string, responseBody []byte, success bool, errorCode string, duration int64, providerName string) {
	log := logger.FromContext(ctx)

	usageLog := &biz.UsageLog{
		RequestID: requestID,
		Provider:  providerName,
		Model:     model,
		Success:   success,
		ErrorCode: errorCode,
		Duration:  duration,
		CreatedAt: time.Now(),
	}

	// 如果请求成功，从响应体中提取 token 使用信息
	if success && len(responseBody) > 0 {
		var respData struct {
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal(responseBody, &respData); err != nil {
			log.Warn("failed to parse usage from response",
				"error", err,
				"request_id", requestID,
			)
		} else {
			usageLog.PromptTokens = respData.Usage.PromptTokens
			usageLog.CompletionTokens = respData.Usage.CompletionTokens
			usageLog.TotalTokens = respData.Usage.TotalTokens

			// 记录 token 使用的 metrics
			metrics.RecordTokenUsage(providerName, model, respData.Usage.PromptTokens, respData.Usage.CompletionTokens)
		}
	}

	// 异步记录，不阻塞主流程
	go func() {
		// 使用新的 context 避免请求取消影响日志记录
		recordCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.usageRepo.Create(recordCtx, usageLog); err != nil {
			log.Error("failed to record usage log",
				"error", err,
				"request_id", requestID,
			)
		}
	}()
}
