package provider

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/config"
	"github.com/acnoway/litellm-go-gateway/internal/data"
)

// Manager 负责 provider 的自动发现、实例化和选择。
// 使用 ModelRouter 根据模型名将请求路由到对应的 provider。
type Manager struct {
	registry *Registry
	router   *DeploymentRouter
}

// NewManager 从配置自动装配所有可用 provider，并构建注册表和路由器。
// 至少需要一个 provider 才能启动；未来可放宽此限制以支持纯健康检查模式。
func NewManager(cfg config.Config, db *sql.DB) (*Manager, error) {
	providers := BuildAll(cfg)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	registry, err := NewRegistry(providers...)
	if err != nil {
		return nil, fmt.Errorf("build provider registry: %w", err)
	}

	// 创建 DeploymentRepo 用于从数据库加载路由规则
	repo := data.NewDeploymentRepo(db)
	// 创建模型路由器，使用第一个 provider 作为默认 fallback
	router := NewDeploymentRouter(repo, registry, providers[0])

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
	return "deployment-router"
}

// GetProvidersForModel 返回指定模型的所有可用 providers（主 + fallbacks）
func (m *Manager) GetProvidersForModel(model string) []biz.Provider {
	return m.router.Route(model)
}

// Registry 暴露底层注册表，供未来多 provider 路由逻辑使用。
func (m *Manager) Registry() *Registry {
	return m.registry
}
