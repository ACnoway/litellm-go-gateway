package httpapi

import (
	"errors"
	"io"
	"net/http"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler 是 OpenAI-compatible HTTP 协议层。它负责 JSON/SSE 和 HTTP 状态码，
// 但不包含 provider 选择、请求转换或上游网络调用等业务职责。
type Handler struct {
	chatService   *service.ChatService
	adminService  *service.AdminService
	deploymentService *service.DeploymentService
}

// NewHandler 通过依赖注入接收聊天用例，令 Handler 可在测试中替换 service。
func NewHandler(chatService *service.ChatService, adminService *service.AdminService, deploymentService *service.DeploymentService) *Handler {
	return &Handler{
		chatService:       chatService,
		adminService:      adminService,
		deploymentService: deploymentService,
	}
}

// Register 注册当前 Gateway 支持的公开路由。管理 API 与其他 OpenAI endpoint
// 应随功能实现增加在此处，保持对外路由集中可见。
func (h *Handler) Register(router gin.IRoutes) {
	router.GET("/healthz", h.health)
	router.GET("/readyz", h.health)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/v1/models", h.models)
	router.POST("/v1/chat/completions", h.chatCompletions)
}

// health 用于 liveness/readiness 探针。当前没有外部依赖的启动检查，
// 因此健康和就绪都以进程能够响应 HTTP 请求为准。
func (h *Handler) health(c *gin.Context) {
	c.Status(http.StatusOK)
}

// models 返回所有可用的逻辑模型列表（从 deployment 读取）。
func (h *Handler) models(c *gin.Context) {
	models, err := h.adminService.ListModels()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "api_error", "failed_to_list_models", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}

// chatCompletions 将 JSON 请求绑定到内部模型，之后按 stream 字段选择响应路径。
func (h *Handler) chatCompletions(c *gin.Context) {
	var request biz.ChatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		// 对输入校验失败返回 OpenAI 风格错误体，而不是 Gin 默认的纯文本错误。
		writeError(c, http.StatusBadRequest, "invalid_request_error", "invalid_request", err.Error())
		return
	}

	if request.Stream {
		h.stream(c, request)
		return
	}
	response, err := h.chatService.Complete(c.Request.Context(), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	// OpenAI adapter 已返回 JSON；直接写字节可避免不必要的反序列化和重编码。
	c.Data(http.StatusOK, "application/json", response.Body)
}

// stream 将上游 SSE 字节原样转发给客户端。原样代理适用于 OpenAI provider；
// 当接入非 OpenAI 协议的 provider 时，应由 adapter 把它转换为 OpenAI SSE 事件。
func (h *Handler) stream(c *gin.Context, request biz.ChatRequest) {
	stream, err := h.chatService.CompleteStream(c.Request.Context(), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	defer stream.Body.Close()

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/event-stream")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	// 每次读取后都 Flush，避免 Web 服务器或反向代理积累多个 chunk 才发送，
	// 这直接影响首 token 延迟。32 KiB 只是读缓冲大小，不会改变 SSE 事件内容。
	buffer := make([]byte, 32*1024)
	for {
		read, readErr := stream.Body.Read(buffer)
		if read > 0 {
			if _, writeErr := c.Writer.Write(buffer[:read]); writeErr != nil {
				// 通常代表客户端已断开；请求 context 会使上游读取尽快结束。
				return
			}
			c.Writer.Flush()
		}
		if errors.Is(readErr, io.EOF) {
			return
		}
		if readErr != nil {
			return
		}
	}
}

// writeServiceError 将 adapter 层已知错误转换为兼容的 HTTP 响应，
// 未分类错误则隐藏实现细节并统一作为 502 返回。
func writeServiceError(c *gin.Context, err error) {
	var providerError *biz.ProviderError
	if errors.As(err, &providerError) {
		writeError(c, providerError.Status, "provider_error", providerError.Code, providerError.Message)
		return
	}
	writeError(c, http.StatusBadGateway, "api_error", "upstream_error", "the upstream provider request failed")
}

// writeError 集中定义 OpenAI 风格的错误结构，保证所有 Handler 分支一致。
func writeError(c *gin.Context, status int, errorType string, code string, message string) {
	c.JSON(status, gin.H{"error": gin.H{"message": message, "type": errorType, "code": code}})
}
