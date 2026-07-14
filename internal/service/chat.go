package service

import (
	"context"

	"github.com/example/litellm-go-gateway/internal/biz"
)

type ChatService struct {
	provider biz.Provider
}

func NewChatService(provider biz.Provider) *ChatService {
	return &ChatService{provider: provider}
}

func (s *ChatService) Complete(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	return s.provider.Chat(ctx, request)
}

func (s *ChatService) CompleteStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	return s.provider.ChatStream(ctx, request)
}
