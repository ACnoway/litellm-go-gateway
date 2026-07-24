package provider

import (
	"context"
	"fmt"
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

// mockDeploymentRepo 是用于测试的模拟 deployment repo
type mockDeploymentRepo struct {
	deployments map[string]*biz.Deployment
}

func (r *mockDeploymentRepo) List(ctx context.Context) ([]*biz.Deployment, error) {
	result := make([]*biz.Deployment, 0, len(r.deployments))
	for _, d := range r.deployments {
		result = append(result, d)
	}
	return result, nil
}

func (r *mockDeploymentRepo) ListEnabled(ctx context.Context) ([]*biz.Deployment, error) {
	return r.List(ctx)
}

func (r *mockDeploymentRepo) Get(ctx context.Context, id int64) (*biz.Deployment, error) {
	for _, d := range r.deployments {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("deployment not found: id=%d", id)
}

func (r *mockDeploymentRepo) GetByName(ctx context.Context, name string) (*biz.Deployment, error) {
	if d, exists := r.deployments[name]; exists {
		return d, nil
	}
	return nil, fmt.Errorf("deployment not found: name=%s", name)
}

func (r *mockDeploymentRepo) Create(ctx context.Context, d *biz.Deployment) error {
	if _, exists := r.deployments[d.Name]; exists {
		return fmt.Errorf("deployment already exists: name=%s", d.Name)
	}
	r.deployments[d.Name] = d
	return nil
}

func (r *mockDeploymentRepo) Update(ctx context.Context, d *biz.Deployment) error {
	if _, exists := r.deployments[d.Name]; !exists {
		return fmt.Errorf("deployment not found: name=%s", d.Name)
	}
	r.deployments[d.Name] = d
	return nil
}

func (r *mockDeploymentRepo) Delete(ctx context.Context, id int64) error {
	for name, d := range r.deployments {
		if d.ID == id {
			delete(r.deployments, name)
			return nil
		}
	}
	return fmt.Errorf("deployment not found: id=%d", id)
}

// mockRegistry 是用于测试的模拟 registry（实现 ProviderRegistry 接口）
type mockRegistry struct {
	providers map[string]biz.Provider
}

func (r *mockRegistry) Get(name string) (biz.Provider, bool) {
	p, exists := r.providers[name]
	return p, exists
}

func (r *mockRegistry) All() []biz.Provider {
	providers := make([]biz.Provider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

func TestDeploymentRouter_Route(t *testing.T) {
	openai := &mockProvider{name: "openai"}
	anthropic := &mockProvider{name: "anthropic"}

	registry := &mockRegistry{
		providers: map[string]biz.Provider{
			"openai":    openai,
			"anthropic": anthropic,
		},
	}

	repo := &mockDeploymentRepo{
		deployments: map[string]*biz.Deployment{
			"gpt-4": {
				Name:      "gpt-4",
				Providers: []string{"openai"},
			},
			"claude-3": {
				Name:      "claude-3",
				Providers: []string{"anthropic"},
			},
			"claude-3-opus": {
				Name:      "claude-3-opus",
				Providers: []string{"anthropic"},
			},
		},
	}

	router := NewDeploymentRouter(repo, registry, openai)

	tests := []struct {
		model             string
		expectedProviders []string
	}{
		{"gpt-4", []string{"openai"}},
		{"claude-3", []string{"anthropic"}},
		{"claude-3-opus", []string{"anthropic"}},
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

func TestDeploymentRouter_GetPrimaryProvider(t *testing.T) {
	openai := &mockProvider{name: "openai"}
	anthropic := &mockProvider{name: "anthropic"}

	registry := &mockRegistry{
		providers: map[string]biz.Provider{
			"openai":    openai,
			"anthropic": anthropic,
		},
	}

	repo := &mockDeploymentRepo{
		deployments: map[string]*biz.Deployment{
			"gpt-4": {
				Name:      "gpt-4",
				Providers: []string{"openai"},
			},
			"claude-3": {
				Name:      "claude-3",
				Providers: []string{"anthropic"},
			},
			"claude-3-opus": {
				Name:      "claude-3-opus",
				Providers: []string{"anthropic"},
			},
		},
	}

	router := NewDeploymentRouter(repo, registry, openai)

	tests := []struct {
		model            string
		expectedProvider string
	}{
		{"gpt-4", "openai"},
		{"claude-3", "anthropic"},
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

func TestDeploymentRouter_GetFallbackProviders(t *testing.T) {
	openai := &mockProvider{name: "openai"}
	anthropic := &mockProvider{name: "anthropic"}
	azure := &mockProvider{name: "azure"}

	registry := &mockRegistry{
		providers: map[string]biz.Provider{
			"openai":    openai,
			"anthropic": anthropic,
			"azure":     azure,
		},
	}

	repo := &mockDeploymentRepo{
		deployments: map[string]*biz.Deployment{
			"gpt-4": {
				Name:      "gpt-4",
				Providers: []string{"openai", "azure", "anthropic"},
			},
			"claude-3": {
				Name:      "claude-3",
				Providers: []string{"anthropic"},
			},
			"claude-3-opus": {
				Name:      "claude-3-opus",
				Providers: []string{"anthropic"},
			},
		},
	}

	router := NewDeploymentRouter(repo, registry, openai)

	tests := []struct {
		model                 string
		expectedFallbackCount int
		expectedFallbackNames []string
	}{
		{"gpt-4", 2, []string{"azure", "anthropic"}},
		{"claude-3", 0, nil},
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

func TestDeploymentRouter_NoFallback(t *testing.T) {
	registry := &mockRegistry{
		providers: map[string]biz.Provider{},
	}

	repo := &mockDeploymentRepo{
		deployments: map[string]*biz.Deployment{},
	}

	router := NewDeploymentRouter(repo, registry, nil)

	providers := router.Route("any-model")
	if providers != nil {
		t.Fatalf("expected nil providers when no fallback is set, got %d providers", len(providers))
	}

	primary := router.GetPrimaryProvider("any-model")
	if primary != nil {
		t.Fatal("expected nil primary provider when no fallback is set")
	}
}
