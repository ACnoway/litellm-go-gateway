package biz

import (
	"context"
	"encoding/json"
	"io"
)

type ChatRequest struct {
	Model       string          `json:"model" binding:"required"`
	Messages    []Message       `json:"messages" binding:"required,min=1"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Tools       json.RawMessage `json:"tools,omitempty"`
}

type Message struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ChatResponse struct {
	Body []byte
}

type ChatStream struct {
	Body io.ReadCloser
}

type Provider interface {
	Name() string
	Chat(context.Context, ChatRequest) (ChatResponse, error)
	ChatStream(context.Context, ChatRequest) (ChatStream, error)
}

type ProviderError struct {
	Status  int
	Code    string
	Message string
}

func (e *ProviderError) Error() string {
	return e.Message
}
