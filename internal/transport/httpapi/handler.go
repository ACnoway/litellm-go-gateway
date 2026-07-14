package httpapi

import (
	"errors"
	"io"
	"net/http"

	"github.com/example/litellm-go-gateway/internal/biz"
	"github.com/example/litellm-go-gateway/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	chatService *service.ChatService
}

func NewHandler(chatService *service.ChatService) *Handler {
	return &Handler{chatService: chatService}
}

func (h *Handler) Register(router gin.IRoutes) {
	router.GET("/healthz", h.health)
	router.GET("/readyz", h.health)
	router.GET("/v1/models", h.models)
	router.POST("/v1/chat/completions", h.chatCompletions)
}

func (h *Handler) health(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (h *Handler) models(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   []gin.H{{"id": "gpt-4o", "object": "model", "owned_by": "openai"}},
	})
}

func (h *Handler) chatCompletions(c *gin.Context) {
	var request biz.ChatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
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
	c.Data(http.StatusOK, "application/json", response.Body)
}

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

	buffer := make([]byte, 32*1024)
	for {
		read, readErr := stream.Body.Read(buffer)
		if read > 0 {
			if _, writeErr := c.Writer.Write(buffer[:read]); writeErr != nil {
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

func writeServiceError(c *gin.Context, err error) {
	var providerError *biz.ProviderError
	if errors.As(err, &providerError) {
		writeError(c, providerError.Status, "provider_error", providerError.Code, providerError.Message)
		return
	}
	writeError(c, http.StatusBadGateway, "api_error", "upstream_error", "the upstream provider request failed")
}

func writeError(c *gin.Context, status int, errorType string, code string, message string) {
	c.JSON(status, gin.H{"error": gin.H{"message": message, "type": errorType, "code": code}})
}
