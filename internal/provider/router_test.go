package provider

import (
	"context"
	"testing"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// mockProvider 是用于测试的模拟 provider
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Chat(ctx context.Context, req biz.ChatRequest) (biz.ChatResponse, error) {
	return biz.ChatResponse{}, nil
}

func (m *mockProvider) ChatStream(ctx context.Context, req biz.ChatRequest) (biz.ChatStream, error) {
	return biz.ChatStream{}, nil
}

func TestModelRouter_AddRoute(t *testing.T) {
	openai := &mockProvider{name: "openai"}
	anthropic := &mockProvider{name: "anthropic"}

	router := NewModelRouter(openai)

	// 添加路由规则
	err := router.AddRoute("^gpt-.*", openai)
	if err != nil {
		t.Fatalf("failed to add route: %v", err)
	}

	err = router.AddRoute("^claude-.*", anthropic)
	if err != nil {
		t.Fatalf("failed to add route: %v", err)
	}

	// 测试空 providers 列表
	err = router.AddRoute("^test-.*")
	if err == nil {
		t.Fatal("expected error for empty providers list")
	}

	// 测试无效的正则表达式
	err = router.AddRoute("[invalid", openai)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestModelRouter_Route(t *testing.T) {
	openai := &mockProvider{name: "openai"}
	anthropic := &mockProvider{name: "anthropic"}
	azure := &mockProvider{name: "azure"}

	router := NewModelRouter(openai)

	// 添加路由规则
	_ = router.AddRoute("^gpt-.*", openai, azure)
	_ = router.AddRoute("^claude-.*", anthropic)

	tests := []struct {
		model             string
		expectedProviders []string
	}{
		{"gpt-4", []string{"openai", "azure"}},
		{"gpt-3.5-turbo", []string{"openai", "azure"}},
		{"claude-3-opus", []string{"anthropic"}},
		{"claude-2", []string{"anthropic"}},
		{"unknown-model", []string{"openai"}}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			providers := router.Route(tt.model)
			if len(providers) != len(tt.expectedProviders) {
				t.Fatalf("expected %d providers, got %d", len(tt.expectedProviders), len(providers))
			}

			for i, expected := range tt.expectedProviders {
				if providers[i].Name() != expected {
					t.Errorf("expected provider %s at index %d, got %s", expected, i, providers[i].Name())
				}
			}
		})
	}
}

func TestModelRouter_GetPrimaryProvider(t *testing.T) {
	openai := &mockProvider{name: "openai"}
	anthropic := &mockProvider{name: "anthropic"}
	azure := &mockProvider{name: "azure"}

	router := NewModelRouter(openai)

	_ = router.AddRoute("^gpt-.*", openai, azure)
	_ = router.AddRoute("^claude-.*", anthropic)

	tests := []struct {
		model            string
		expectedProvider string
	}{
		{"gpt-4", "openai"},
		{"claude-3-opus", "anthropic"},
		{"unknown-model", "openai"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			provider := router.GetPrimaryProvider(tt.model)
			if provider == nil {
				t.Fatal("expected non-nil provider")
			}
			if provider.Name() != tt.expectedProvider {
				t.Errorf("expected provider %s, got %s", tt.expectedProvider, provider.Name())
			}
		})
	}
}

func TestModelRouter_GetFallbackProviders(t *testing.T) {
	openai := &mockProvider{name: "openai"}
	anthropic := &mockProvider{name: "anthropic"}
	azure := &mockProvider{name: "azure"}

	router := NewModelRouter(openai)

	_ = router.AddRoute("^gpt-.*", openai, azure, anthropic)
	_ = router.AddRoute("^claude-.*", anthropic)

	tests := []struct {
		model                   string
		expectedFallbackCount   int
		expectedFallbackNames   []string
	}{
		{"gpt-4", 2, []string{"azure", "anthropic"}},
		{"claude-3-opus", 0, nil},
		{"unknown-model", 0, nil}, // fallback 只有一个 provider
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			fallbacks := router.GetFallbackProviders(tt.model)
			if len(fallbacks) != tt.expectedFallbackCount {
				t.Fatalf("expected %d fallback providers, got %d", tt.expectedFallbackCount, len(fallbacks))
			}

			for i, expected := range tt.expectedFallbackNames {
				if fallbacks[i].Name() != expected {
					t.Errorf("expected fallback provider %s at index %d, got %s", expected, i, fallbacks[i].Name())
				}
			}
		})
	}
}

func TestModelRouter_NoFallback(t *testing.T) {
	router := NewModelRouter(nil)

	providers := router.Route("any-model")
	if providers != nil {
		t.Fatalf("expected nil providers when no fallback is set, got %d providers", len(providers))
	}

	primary := router.GetPrimaryProvider("any-model")
	if primary != nil {
		t.Fatal("expected nil primary provider when no fallback is set")
	}
}
