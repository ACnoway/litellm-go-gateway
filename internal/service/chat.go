package service

import (
	"context"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// ChatService 是聊天用例的编排层。目前它只把调用委托给单个 provider；
// 后续鉴权结果、模型 deployment 路由、重试、fallback 和用量统计都应加入此层，
// 而非散落到 Gin Handler 或具体的 provider adapter。
type ChatService struct {
	provider biz.Provider
}

// NewChatService 通过接口注入 provider，便于替换实现和单元测试。
func NewChatService(provider biz.Provider) *ChatService {
	return &ChatService{provider: provider}
}

// Complete 执行非流式聊天调用。ctx 来自 HTTP 请求，客户端取消时会传递到上游请求。
func (s *ChatService) Complete(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	return s.provider.Chat(ctx, request)
}

// CompleteStream 执行流式聊天调用，并把仍打开的上游流交给 Handler 转发。
func (s *ChatService) CompleteStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	return s.provider.ChatStream(ctx, request)
}
