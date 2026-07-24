package service

import (
	"context"
	"fmt"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/provider"
)

// AdminService 提供管理接口的业务逻辑
type AdminService struct {
	providerManager *provider.Manager
	routingRepo     biz.RoutingRuleRepo
	deploymentRepo  biz.DeploymentRepo
}

// NewAdminService 创建管理服务
func NewAdminService(providerManager *provider.Manager, routingRepo biz.RoutingRuleRepo, deploymentRepo biz.DeploymentRepo) *AdminService {
	return &AdminService{
		providerManager: providerManager,
		routingRepo:     routingRepo,
		deploymentRepo:  deploymentRepo,
	}
}

// ListProviders 返回所有已注册的 provider 信息
func (s *AdminService) ListProviders() []biz.ProviderInfo {
	registry := s.providerManager.Registry()
	providers := registry.All()

	result := make([]biz.ProviderInfo, 0, len(providers))
	for _, p := range providers {
		if p == nil {
			continue
		}
		result = append(result, biz.ProviderInfo{
			Name:      p.Name(),
			Available: true,
			Type:      p.Name(),
		})
	}
	return result
}

// GetProvider 返回指定 provider 的信息
func (s *AdminService) GetProvider(name string) (*biz.ProviderInfo, error) {
	registry := s.providerManager.Registry()
	p, exists := registry.Get(name)
	if !exists || p == nil {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return &biz.ProviderInfo{
		Name:      p.Name(),
		Available: true,
		Type:      p.Name(),
	}, nil
}

// ListRoutingRules 返回所有路由规则
func (s *AdminService) ListRoutingRules() ([]biz.RoutingRuleResponse, error) {
	return s.routingRepo.List()
}

// GetRoutingRule 返回单条路由规则
func (s *AdminService) GetRoutingRule(id int) (*biz.RoutingRuleResponse, error) {
	return s.routingRepo.Get(id)
}

// CreateRoutingRule 创建新的路由规则
func (s *AdminService) CreateRoutingRule(req biz.RoutingRuleRequest) (*biz.RoutingRuleResponse, error) {
	// 验证 providers 是否存在
	registry := s.providerManager.Registry()
	for _, providerName := range req.Providers {
		if p, exists := registry.Get(providerName); !exists || p == nil {
			return nil, fmt.Errorf("provider not found: %s", providerName)
		}
	}

	// 验证 pattern 是否为有效的正则表达式
	if err := validatePattern(req.Pattern); err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	return s.routingRepo.Create(req.Pattern, req.Providers)
}

// UpdateRoutingRule 更新路由规则
func (s *AdminService) UpdateRoutingRule(id int, req biz.RoutingRuleRequest) (*biz.RoutingRuleResponse, error) {
	// 验证 providers 是否存在
	registry := s.providerManager.Registry()
	for _, providerName := range req.Providers {
		if p, exists := registry.Get(providerName); !exists || p == nil {
			return nil, fmt.Errorf("provider not found: %s", providerName)
		}
	}

	// 验证 pattern 是否为有效的正则表达式
	if err := validatePattern(req.Pattern); err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	return s.routingRepo.Update(id, req.Pattern, req.Providers)
}

// DeleteRoutingRule 删除路由规则
func (s *AdminService) DeleteRoutingRule(id int) error {
	return s.routingRepo.Delete(id)
}

// validatePattern 验证路由规则的 pattern 是否有效
func validatePattern(pattern string) error {
	// 这里可以添加更复杂的验证逻辑
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	return nil
}

// ListModels 返回所有可用的逻辑模型列表（用于 /v1/models 端点）
func (s *AdminService) ListModels() ([]biz.ModelInfo, error) {
	deployments, err := s.deploymentRepo.ListEnabled(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	result := make([]biz.ModelInfo, 0, len(deployments))
	for _, d := range deployments {
		// 使用第一个 provider 作为 owned_by
		ownedBy := "unknown"
		if len(d.Providers) > 0 {
			ownedBy = d.Providers[0]
		}

		result = append(result, biz.ModelInfo{
			ID:          d.Name,
			Object:      "model",
			Created:     d.CreatedAt.Unix(),
			OwnedBy:     ownedBy,
			Ready:       d.Enabled,
			Description: d.Description,
			Providers:   d.Providers,
		})
	}

	return result, nil
}

