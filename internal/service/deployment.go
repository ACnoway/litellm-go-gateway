package service

import (
	"context"
	"fmt"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	"github.com/acnoway/litellm-go-gateway/internal/pkg/logger"
)

// DeploymentService 处理 deployment 的业务逻辑。
type DeploymentService struct {
	repo biz.DeploymentRepo
}

// NewDeploymentService 创建 DeploymentService。
func NewDeploymentService(repo biz.DeploymentRepo) *DeploymentService {
	return &DeploymentService{repo: repo}
}

// ListDeployments 返回所有 deployments。
func (s *DeploymentService) ListDeployments(ctx context.Context) ([]*biz.Deployment, error) {
	log := logger.FromContext(ctx)
	deployments, err := s.repo.List(ctx)
	if err != nil {
		log.Error("failed to list deployments", "error", err)
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	return deployments, nil
}

// GetDeployment 根据 ID 获取单个 deployment。
func (s *DeploymentService) GetDeployment(ctx context.Context, id int64) (*biz.Deployment, error) {
	log := logger.FromContext(ctx)
	deployment, err := s.repo.Get(ctx, id)
	if err != nil {
		log.Error("failed to get deployment", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	return deployment, nil
}

// CreateDeployment 创建新的 deployment。
func (s *DeploymentService) CreateDeployment(ctx context.Context, req biz.DeploymentRequest) (*biz.Deployment, error) {
	log := logger.FromContext(ctx)

	// 验证请求
	if err := s.validateDeploymentRequest(ctx, req); err != nil {
		return nil, err
	}

	// 设置默认值
	if req.Strategy == "" {
		req.Strategy = "priority"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	deployment := &biz.Deployment{
		Name:        req.Name,
		ActualModel: req.ActualModel,
		Providers:   req.Providers,
		Strategy:    req.Strategy,
		Weights:     req.Weights,
		MaxTokens:   req.MaxTokens,
		Description: req.Description,
		Enabled:     enabled,
	}

	if err := s.repo.Create(ctx, deployment); err != nil {
		log.Error("failed to create deployment", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	log.Info("deployment created", "id", deployment.ID, "name", deployment.Name)
	return deployment, nil
}

// UpdateDeployment 更新已有的 deployment。
func (s *DeploymentService) UpdateDeployment(ctx context.Context, id int64, req biz.DeploymentRequest) (*biz.Deployment, error) {
	log := logger.FromContext(ctx)

	// 验证请求
	if err := s.validateDeploymentRequest(ctx, req); err != nil {
		return nil, err
	}

	// 获取现有 deployment
	deployment, err := s.repo.Get(ctx, id)
	if err != nil {
		log.Error("failed to get deployment for update", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// 更新字段
	deployment.Name = req.Name
	deployment.ActualModel = req.ActualModel
	deployment.Providers = req.Providers
	if req.Strategy != "" {
		deployment.Strategy = req.Strategy
	}
	deployment.Weights = req.Weights
	deployment.MaxTokens = req.MaxTokens
	deployment.Description = req.Description
	if req.Enabled != nil {
		deployment.Enabled = *req.Enabled
	}

	if err := s.repo.Update(ctx, deployment); err != nil {
		log.Error("failed to update deployment", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update deployment: %w", err)
	}

	log.Info("deployment updated", "id", deployment.ID, "name", deployment.Name)
	return deployment, nil
}

// DeleteDeployment 删除 deployment。
func (s *DeploymentService) DeleteDeployment(ctx context.Context, id int64) error {
	log := logger.FromContext(ctx)

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error("failed to delete deployment", "id", id, "error", err)
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	log.Info("deployment deleted", "id", id)
	return nil
}

// validateDeploymentRequest 验证 deployment 请求的合法性。
func (s *DeploymentService) validateDeploymentRequest(ctx context.Context, req biz.DeploymentRequest) error {
	// 验证 strategy
	validStrategies := map[string]bool{
		"priority":    true,
		"round-robin": true,
		"weighted":    true,
	}
	if req.Strategy != "" && !validStrategies[req.Strategy] {
		return fmt.Errorf("invalid strategy: %s (valid values: priority, round-robin, weighted)", req.Strategy)
	}

	// 如果是 weighted 策略，验证 weights 长度与 providers 一致
	if req.Strategy == "weighted" {
		if len(req.Weights) != len(req.Providers) {
			return fmt.Errorf("weights length must match providers length for weighted strategy")
		}
		// 验证权重都是正数
		for i, weight := range req.Weights {
			if weight <= 0 {
				return fmt.Errorf("weight at index %d must be positive", i)
			}
		}
	}

	return nil
}
