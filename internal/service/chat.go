package service

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
)

// ChatService 是聊天用例的编排层。目前它只把调用委托给单个 provider；
// 后续鉴权结果、模型 deployment 路由、重试、fallback 和用量统计都应加入此层，
// 而非散落到 Gin Handler 或具体的 provider adapter。
type ChatService struct {
	provider    biz.Provider
	retryConfig config.RetryConfig
}

// NewChatService 通过接口注入 provider，便于替换实现和单元测试。
func NewChatService(provider biz.Provider, retryConfig config.RetryConfig) *ChatService {
	return &ChatService{
		provider:    provider,
		retryConfig: retryConfig,
	}
}

// Complete 执行非流式聊天调用。ctx 来自 HTTP 请求，客户端取消时会传递到上游请求。
func (s *ChatService) Complete(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	return s.withRetry(ctx, func() (biz.ChatResponse, error) {
		return s.provider.Chat(ctx, request)
	})
}

// CompleteStream 执行流式聊天调用，并把仍打开的上游流交给 Handler 转发。
// 流式请求不重试，因为响应体已经开始传输，中途重试会导致客户端收到重复或错误的数据。
func (s *ChatService) CompleteStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	return s.provider.ChatStream(ctx, request)
}

// withRetry 实现指数退避重试逻辑。只对网络错误（连接失败、超时、DNS 解析失败）重试，
// 对于 4xx 客户端错误或 5xx 服务端错误不重试，因为立即重试不太可能成功。
func (s *ChatService) withRetry(ctx context.Context, fn func() (biz.ChatResponse, error)) (biz.ChatResponse, error) {
	var lastErr error
	delay := s.retryConfig.InitialDelay

	for attempt := 1; attempt <= s.retryConfig.MaxAttempts; attempt++ {
		resp, err := fn()
		if err == nil {
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
			break
		}

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
