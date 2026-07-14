package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// Provider 是 OpenAI Chat Completions API 的 adapter。它把 Gateway 的内部请求
// 转为 OpenAI HTTP 请求，并将上游错误规范化为 biz.ProviderError。
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New 创建可复用的 OpenAI adapter。TrimRight 防止用户把 baseURL 配为
// ".../v1/" 时与后续拼接的 "/chat/completions" 形成双斜杠。
func New(apiKey string, baseURL string, client *http.Client) *Provider {
	return &Provider{apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

// Name 是注册表和未来路由器使用的稳定 provider 标识。
func (p *Provider) Name() string {
	return "openai"
}

// Chat 用于非流式调用。它读取并关闭完整响应体，调用方只得到已完成的 JSON 字节。
func (p *Provider) Chat(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	response, err := p.do(ctx, request)
	if err != nil {
		return biz.ChatResponse{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return biz.ChatResponse{}, fmt.Errorf("read OpenAI response: %w", err)
	}
	return biz.ChatResponse{Body: body}, nil
}

// ChatStream 用于 SSE 调用。此处不能读取或关闭响应体，因为 Handler 还需将其
// 增量写给客户端；响应体的关闭责任随 ChatStream 一并转移到 Handler。
func (p *Provider) ChatStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	response, err := p.do(ctx, request)
	if err != nil {
		return biz.ChatStream{}, err
	}
	return biz.ChatStream{Body: response.Body}, nil
}

// do 是两种调用共享的上游请求流程。请求 context 从 HTTP 入站请求一路传递，
// 因而客户端取消、Gateway 超时或服务关机时，标准库会中止对应的上游连接。
func (p *Provider) do(ctx context.Context, request biz.ChatRequest) (*http.Response, error) {
	if p.apiKey == "" {
		return nil, &biz.ProviderError{Status: http.StatusServiceUnavailable, Code: "provider_not_configured", Message: "OpenAI API key is not configured"}
	}

	// 首期内部请求模型与 OpenAI 格式相容，可以直接编码。接入非兼容 provider 时，
	// 应在其 adapter 中转换为供应商格式，而不是污染 biz.ChatRequest。
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode OpenAI request: %w", err)
	}
	upstreamRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build OpenAI request: %w", err)
	}
	upstreamRequest.Header.Set("Authorization", "Bearer "+p.apiKey)
	upstreamRequest.Header.Set("Content-Type", "application/json")
	// 同时接受 JSON 和 SSE，使同一 adapter 能为 Stream=false/true 服务。
	upstreamRequest.Header.Set("Accept", "application/json, text/event-stream")

	response, err := p.client.Do(upstreamRequest)
	if err != nil {
		return nil, fmt.Errorf("call OpenAI: %w", err)
	}
	if response.StatusCode < http.StatusBadRequest {
		return response, nil
	}
	defer response.Body.Close()

	return nil, parseProviderError(response)
}

// parseProviderError 优先读取 OpenAI 标准 error 对象，再回退为原始响应文本。
// 限制读取为 1 MiB，防止异常上游返回巨大错误体而耗尽 Gateway 内存。
func parseProviderError(response *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return &biz.ProviderError{Status: response.StatusCode, Code: "provider_error", Message: response.Status}
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.Error.Message != "" {
		return &biz.ProviderError{Status: response.StatusCode, Code: payload.Error.Code, Message: payload.Error.Message}
	}
	return &biz.ProviderError{Status: response.StatusCode, Code: "provider_error", Message: strings.TrimSpace(string(body))}
}
