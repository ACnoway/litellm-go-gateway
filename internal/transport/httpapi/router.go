package httpapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewRouter 组装 Gin 路由与全局中间件。中间件注册顺序就是执行顺序：
// panic 恢复最外层，随后写入请求 ID，最后才执行受保护接口的鉴权。
func NewRouter(handler *Handler, gatewayAPIKey string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), requestID(), authorize(gatewayAPIKey))
	handler.Register(router)
	return router
}

// requestID 优先沿用调用方提供的 X-Request-ID；不存在时生成 128 位随机 ID。
// 返回该 header 方便调用方将 Gateway 日志与自身请求日志关联。
func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Header("X-Request-ID", requestID)
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
func authorize(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" || c.Request.URL.Path == "/healthz" || c.Request.URL.Path == "/readyz" {
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
