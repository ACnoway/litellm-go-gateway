package biz

import (
	"context"
	"encoding/json"
	"io"
)

// ChatRequest 是 Gateway 的内部聊天请求模型。当前字段与 OpenAI Chat Completions
// 高度接近以便首个 adapter 直接序列化；未来 provider 特有的转换应发生在 provider 层，
// 而不是让 HTTP Handler 感知不同厂商协议。
type ChatRequest struct {
	Model       string          `json:"model" binding:"required"`
	Messages    []Message       `json:"messages" binding:"required,min=1"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Tools       json.RawMessage `json:"tools,omitempty"`
}

// Message 是当前最小化的文本消息模型。Content 暂时只接受字符串；
// OpenAI 的多模态 content part 和 tool message 可在扩展规范化模型时加入。
type Message struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// ChatResponse 保存 provider 返回的非流式 JSON 字节。Handler 不需要再次反序列化，
// 从而保留上游 OpenAI 响应的字段，并减少一次编码开销。
type ChatResponse struct {
	Body []byte
}

// ChatStream 持有仍然打开的上游响应体。所有权转移给 HTTP Handler，
// Handler 在 SSE 转发结束、客户端断开或发生错误时必须关闭它。
type ChatStream struct {
	Body io.ReadCloser
}

// Provider 是每个厂商 adapter 必须实现的端口。Service 只依赖此接口，
// 因此增加 provider 或加入路由选择器不会改变上层 HTTP API。
type Provider interface {
	Name() string
	Chat(context.Context, ChatRequest) (ChatResponse, error)
	ChatStream(context.Context, ChatRequest) (ChatStream, error)
}

// ProviderError 以统一形式表示可安全暴露给调用方的上游错误。
// Status 用于 HTTP 状态码，Code 便于客户端程序化处理，Message 是用户可读说明。
type ProviderError struct {
	Status  int
	Code    string
	Message string
}

func (e *ProviderError) Error() string {
	return e.Message
}
