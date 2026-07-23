package provider

import (
	"context"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// ProviderRegistry 是 provider 索引接口
type ProviderRegistry interface {
	Get(name string) (biz.Provider, bool)
	All() []biz.Provider
}

// DeploymentRouter 负责根据部署配置将请求路由到合适的 provider
type DeploymentRouter struct {
	repo     biz.DeploymentRepo
	registry ProviderRegistry
	fallback biz.Provider // 默认 provider（当没有匹配的路由时使用）
}

// NewDeploymentRouter 创建一个新的Deployment路由器
func NewDeploymentRouter(repo biz.DeploymentRepo, registry ProviderRegistry, fallback biz.Provider) *DeploymentRouter {
	return &DeploymentRouter{
		repo:     repo,
		registry: registry,
		fallback: fallback,
	}
}

// Route 根据模型名返回匹配的 provider 列表
// 返回第一个匹配的路由，如果没有匹配则返回 fallback provider
func (r *DeploymentRouter) Route(model string) []biz.Provider {
	deployment, err := r.repo.GetByName(context.Background(), model)
	if err != nil {
		if r.fallback != nil {
			return []biz.Provider{r.fallback}
		}
		return nil
	}
	providers := make([]biz.Provider, 0, len(deployment.Providers))
	for _, providerName := range deployment.Providers {
		provider, exists := r.registry.Get(providerName)
		if exists && provider != nil {
			providers = append(providers, provider)
		}
	}
	if len(providers) == 0 {
		if r.fallback != nil {
			return []biz.Provider{r.fallback}
		}
		return nil
	}
	return providers
}

// GetPrimaryProvider 返回指定模型的主 provider（第一个 provider）
func (r *DeploymentRouter) GetPrimaryProvider(model string) biz.Provider {
	providers := r.Route(model)
	if len(providers) > 0 {
		return providers[0]
	}
	return nil
}

// GetFallbackProviders 返回指定模型的 fallback providers（除第一个外的所有 providers）
func (r *DeploymentRouter) GetFallbackProviders(model string) []biz.Provider {
	providers := r.Route(model)
	if len(providers) > 1 {
		return providers[1:]
	}
	return nil
}
