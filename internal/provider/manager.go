package provider

import (
	"context"
	"fmt"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
)

// Manager 负责 provider 的自动发现、实例化和选择。
// 当前使用简单的 fallback 策略；未来可扩展为基于模型名的路由。
type Manager struct {
	registry *Registry
	primary  biz.Provider
}

// NewManager 从配置自动装配所有可用 provider，并构建注册表。
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

	return &Manager{
		registry: registry,
		primary:  providers[0], // 当前默认使用第一个；未来可基于优先级排序
	}, nil
}

// Chat 将请求路由到选定的 provider。
// 当前直接使用 primary；未来可根据 request.Model 选择上游。
func (m *Manager) Chat(ctx context.Context, request biz.ChatRequest) (biz.ChatResponse, error) {
	return m.primary.Chat(ctx, request)
}

// ChatStream 将流式请求路由到选定的 provider。
func (m *Manager) ChatStream(ctx context.Context, request biz.ChatRequest) (biz.ChatStream, error) {
	return m.primary.ChatStream(ctx, request)
}

// Name 返回当前选定 provider 的名称（用于日志和监控）。
func (m *Manager) Name() string {
	return m.primary.Name()
}

// Registry 暴露底层注册表，供未来多 provider 路由逻辑使用。
func (m *Manager) Registry() *Registry {
	return m.registry
}