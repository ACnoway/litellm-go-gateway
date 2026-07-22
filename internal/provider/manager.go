package provider

import (
	"context"
	"fmt"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
)

// Manager 负责 provider 的自动发现、实例化和选择。
// 使用 ModelRouter 根据模型名将请求路由到对应的 provider。
type Manager struct {
	registry *Registry
	router   *ModelRouter
}

// NewManager 从配置自动装配所有可用 provider，并构建注册表和路由器。
// 至少需要一个 provider 才能启动；未来可放宽此限制以支持纯健康检查模式。
func NewManager(cfg config.Config) (*Manager, error) {
	providers := BuildAll(cfg)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	registry, err := NewRegistry(providers...)
	if err != nil {
		return nil, fmt.Errorf("build provider registry: %w", err)
	}

	// 创建模型路由器，使用第一个 provider 作为默认 fallback
	router := NewModelRouter(providers[0])

	// 从配置中加载路由规则
	if err := loadRoutingRules(router, cfg, registry); err != nil {
		return nil, fmt.Errorf("load routing rules: %w", err)
	}

	return &Manager{
		registry: registry,
		router:   router,
	}, nil
}

// Chat 根据请求的模型名将请求路由到对应的 provider。
func (m *Manager) Chat(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	provider := m.router.GetPrimaryProvider(request.Model)
	if provider == nil {
		return biz.ChatResponse{}, fmt.Errorf("no provider available for model %s", request.Model)
	}
	return provider.Chat(ctx, request)
}

// ChatStream 根据请求的模型名将流式请求路由到对应的 provider。
func (m *Manager) ChatStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	provider := m.router.GetPrimaryProvider(request.Model)
	if provider == nil {
		return biz.ChatStream{}, fmt.Errorf("no provider available for model %s", request.Model)
	}
	return provider.ChatStream(ctx, request)
}

// Name 返回当前使用的路由器信息（用于日志和监控）。
func (m *Manager) Name() string {
	return "model-router"
}

// GetProvidersForModel 返回指定模型的所有可用 providers（主 + fallbacks）
func (m *Manager) GetProvidersForModel(model string) []biz.Provider {
	return m.router.Route(model)
}

// Registry 暴露底层注册表，供未来多 provider 路由逻辑使用。
func (m *Manager) Registry() *Registry {
	return m.registry
}

// loadRoutingRules 从配置加载模型路由规则
func loadRoutingRules(router *ModelRouter, cfg config.Config, registry *Registry) error {
	// 如果配置中定义了路由规则，则加载它们
	for _, rule := range cfg.Routing.Rules {
		providers := make([]biz.Provider, 0, len(rule.Providers))
		for _, providerName := range rule.Providers {
			provider, exists := registry.Get(providerName)
			if !exists || provider == nil {
				return fmt.Errorf("provider %s not found for model pattern %s", providerName, rule.Pattern)
			}
			providers = append(providers, provider)
		}
		if err := router.AddRoute(rule.Pattern, providers...); err != nil {
			return err
		}
	}

	// 如果没有配置规则，添加默认的智能路由
	if len(cfg.Routing.Rules) == 0 {
		addDefaultRoutes(router, registry)
	}

	return nil
}

// addDefaultRoutes 添加基于 provider 能力的默认路由规则
func addDefaultRoutes(router *ModelRouter, registry *Registry) {
	// OpenAI 模型路由到 OpenAI provider
	if openai, exists := registry.Get("openai"); exists && openai != nil {
		_ = router.AddRoute("^gpt-.*", openai)
		_ = router.AddRoute("^text-.*", openai)
		_ = router.AddRoute("^davinci.*", openai)
	}

	// Anthropic 模型路由到 Anthropic provider
	if anthropic, exists := registry.Get("anthropic"); exists && anthropic != nil {
		_ = router.AddRoute("^claude-.*", anthropic)
	}

	// Azure 模型可以处理 OpenAI 兼容的模型
	// 用户可以通过配置文件覆盖这些默认规则
}