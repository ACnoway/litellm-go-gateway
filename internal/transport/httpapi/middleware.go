package httpapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/acnoway/litellm-go-gateway/internal/pkg/logger"
	"github.com/acnoway/litellm-go-gateway/internal/pkg/metrics"
	"github.com/gin-gonic/gin"
)

// requestID 优先沿用调用方提供的 X-Request-ID；不存在时生成 128 位随机 ID。
// 返回该 header 方便调用方将 Gateway 日志与自身请求日志关联。
// 同时将请求 ID 注入到 context，供后续日志记录使用。
func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Header("X-Request-ID", requestID)

		// 将请求 ID 注入到 context
		ctx := logger.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// newRequestID 用 crypto/rand 而非可预测的伪随机数生成器，避免高并发下冲突，
// 也避免请求 ID 被推测。十六进制编码使其适合 HTTP header 和日志系统。
func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "request-id-unavailable"
	}
	return hex.EncodeToString(value)
}

// authorize 目前实现单个 Gateway API Key。若未设置 GATEWAY_API_KEY，则为了本地开发
// 放行全部请求；生产环境应始终设置它，后续再替换为虚拟 Key/用户/团队鉴权。
// /healthz, /readyz, /metrics 端点无需鉴权，便于监控系统访问。
func authorize(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if apiKey == "" || path == "/healthz" || path == "/readyz" || path == "/metrics" {
			c.Next()
			return
		}

		providedKey := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		// 恒定时间比较避免通过响应耗时推测正确 key 的前缀。
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(apiKey)) == 1 {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "invalid API key", "type": "authentication_error", "code": "invalid_api_key"}})
	}
}

// logging 记录每个请求的基本信息：方法、路径、状态码、响应时间、请求 ID。
// 该中间件应在 requestID 中间件之后执行，以确保请求 ID 已经注入到 context。
func logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 跳过健康检查端点的日志记录，避免日志噪音
		if path == "/healthz" || path == "/readyz" {
			return
		}

		duration := time.Since(start)
		status := c.Writer.Status()

		log := logger.FromContext(c.Request.Context())
		log.Info("request completed",
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

// metricsMiddleware 记录 Prometheus metrics：请求总数、响应时间、错误率。
// 应在 logging 中间件之后执行，以确保所有请求都被记录。
// /metrics 端点本身不记录指标，避免递归。
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 跳过 /metrics 端点本身，避免递归
		if path == "/metrics" {
			return
		}

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		// 记录指标
		metrics.RecordHTTPRequest(path, method, status, duration)
	}
}
