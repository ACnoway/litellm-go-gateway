package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func New(apiKey string, baseURL string, client *http.Client) *Provider {
	return &Provider{apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

func (p *Provider) Name() string {
	return "anthropic"
}

func (p *Provider) Chat(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	response, err := p.do(ctx, request)
	if err != nil {
		return biz.ChatResponse{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return biz.ChatResponse{}, fmt.Errorf("read Anthropic response: %w", err)
	}

	convertedBody, err := p.convertResponseToOpenAIFormat(body)
	if err != nil {
		return biz.ChatResponse{}, fmt.Errorf("convert Anthropic response: %w", err)
	}
	return biz.ChatResponse{Body: convertedBody}, nil
}

func (p *Provider) ChatStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	response, err := p.do(ctx, request)
	if err != nil {
		return biz.ChatStream{}, err
	}
	return biz.ChatStream{Body: &streamConverter{scanner: bufio.NewScanner(response.Body), upstream: response.Body}}, nil
}

func (p *Provider) do(ctx context.Context, request biz.ChatRequest) (*http.Response, error) {
	if p.apiKey == "" {
		return nil, &biz.ProviderError{Status: http.StatusServiceUnavailable, Code: "provider_not_configured", Message: "Anthropic API key is not configured"}
	}

	anthropicReq := p.convertRequestToAnthropicFormat(request)
	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("encode Anthropic request: %w", err)
	}

	upstreamRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build Anthropic request: %w", err)
	}
	upstreamRequest.Header.Set("x-api-key", p.apiKey)
	upstreamRequest.Header.Set("anthropic-version", "2023-06-01")
	upstreamRequest.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(upstreamRequest)
	if err != nil {
		return nil, fmt.Errorf("call Anthropic: %w", err)
	}
	if response.StatusCode < http.StatusBadRequest {
		return response, nil
	}
	defer response.Body.Close()

	return nil, parseProviderError(response)
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float64           `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	System      string             `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *Provider) convertRequestToAnthropicFormat(request biz.ChatRequest) anthropicRequest {
	var system string
	messages := make([]anthropicMessage, 0, len(request.Messages))

	for _, msg := range request.Messages {
		if msg.Role == "system" {
			system = msg.Content
			continue
		}
		messages = append(messages, anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	maxTokens := 4096
	if request.MaxTokens != nil {
		maxTokens = *request.MaxTokens
	}

	return anthropicRequest{
		Model:       request.Model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: request.Temperature,
		Stream:      request.Stream,
		System:      system,
	}
}

func (p *Provider) convertResponseToOpenAIFormat(body []byte) ([]byte, error) {
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, err
	}

	var text string
	if len(anthropicResp.Content) > 0 {
		text = anthropicResp.Content[0].Text
	}

	openAIResp := map[string]interface{}{
		"id":      anthropicResp.ID,
		"object":  "chat.completion",
		"created": 0,
		"model":   anthropicResp.Model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]string{
					"role":    anthropicResp.Role,
					"content": text,
				},
				"finish_reason": anthropicResp.StopReason,
			},
		},
		"usage": map[string]int{
			"prompt_tokens":     anthropicResp.Usage.InputTokens,
			"completion_tokens": anthropicResp.Usage.OutputTokens,
			"total_tokens":      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}

	return json.Marshal(openAIResp)
}

type streamConverter struct {
	scanner  *bufio.Scanner
	upstream io.ReadCloser
}

func (s *streamConverter) Read(p []byte) (n int, err error) {
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return 0, err
		}
		return 0, io.EOF
	}

	line := s.scanner.Text()
	if !strings.HasPrefix(line, "data: ") {
		return s.Read(p)
	}

	data := strings.TrimPrefix(line, "data: ")
	if data == "[DONE]" {
		copy(p, []byte("data: [DONE]\n\n"))
		return len("data: [DONE]\n\n"), io.EOF
	}

	var anthropicChunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &anthropicChunk); err != nil {
		return s.Read(p)
	}

	eventType, _ := anthropicChunk["type"].(string)
	if eventType == "content_block_delta" {
		delta, _ := anthropicChunk["delta"].(map[string]interface{})
		text, _ := delta["text"].(string)

		openAIChunk := map[string]interface{}{
			"id":      "chatcmpl-anthropic",
			"object":  "chat.completion.chunk",
			"created": 0,
			"model":   "claude",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"delta": map[string]string{
						"content": text,
					},
					"finish_reason": nil,
				},
			},
		}

		chunkBytes, _ := json.Marshal(openAIChunk)
		formatted := fmt.Sprintf("data: %s\n\n", chunkBytes)
		copy(p, []byte(formatted))
		return len(formatted), nil
	}

	return s.Read(p)
}

func (s *streamConverter) Close() error {
	return s.upstream.Close()
}

func parseProviderError(response *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return &biz.ProviderError{Status: response.StatusCode, Code: "provider_error", Message: response.Status}
	}

	var payload struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.Error.Message != "" {
		return &biz.ProviderError{Status: response.StatusCode, Code: payload.Error.Type, Message: payload.Error.Message}
	}
	return &biz.ProviderError{Status: response.StatusCode, Code: "provider_error", Message: strings.TrimSpace(string(body))}
}
