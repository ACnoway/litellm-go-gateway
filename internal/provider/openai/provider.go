package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/example/litellm-go-gateway/internal/biz"
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
	return "openai"
}

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

func (p *Provider) ChatStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	response, err := p.do(ctx, request)
	if err != nil {
		return biz.ChatStream{}, err
	}
	return biz.ChatStream{Body: response.Body}, nil
}

func (p *Provider) do(ctx context.Context, request biz.ChatRequest) (*http.Response, error) {
	if p.apiKey == "" {
		return nil, &biz.ProviderError{Status: http.StatusServiceUnavailable, Code: "provider_not_configured", Message: "OpenAI API key is not configured"}
	}

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
