package biz

import (
	"context"
	"time"
)

// Deployment 表示一个逻辑模型的部署配置。
// 它将用户可见的逻辑模型名（如 "gpt-4-turbo"）映射到一个或多个物理 provider 实例，
// 并定义负载均衡策略和模型约束。
type Deployment struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`                    // 逻辑模型名（暴露给用户）
	ActualModel string    `json:"actual_model"`            // 真实模型名（发给上游）
	Providers   []string  `json:"providers"`               // 可用的 providers（按优先级）
	Strategy    string    `json:"strategy"`                // 负载均衡策略: "priority", "round-robin", "weighted"
	Weights     []int     `json:"weights,omitempty"`       // weighted 策略的权重
	MaxTokens   int       `json:"max_tokens,omitempty"`    // 模型的最大 token 限制
	Description string    `json:"description,omitempty"`   // 描述
	Enabled     bool      `json:"enabled"`                 // 是否启用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeploymentRequest 用于创建和更新 Deployment 的请求。
type DeploymentRequest struct {
	Name        string   `json:"name" binding:"required"`
	ActualModel string   `json:"actual_model" binding:"required"`
	Providers   []string `json:"providers" binding:"required,min=1"`
	Strategy    string   `json:"strategy"`                      // 默认 "priority"
	Weights     []int    `json:"weights,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Description string   `json:"description,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`             // 使用指针以区分未设置和 false
}

// DeploymentRepo 定义 Deployment 的数据访问接口。
type DeploymentRepo interface {
	// List 返回所有 deployments。
	List(ctx context.Context) ([]*Deployment, error)

	// ListEnabled 返回所有启用的 deployments。
	ListEnabled(ctx context.Context) ([]*Deployment, error)

	// Get 根据 ID 获取单个 deployment。
	Get(ctx context.Context, id int64) (*Deployment, error)

	// GetByName 根据逻辑模型名获取 deployment。
	GetByName(ctx context.Context, name string) (*Deployment, error)

	// Create 创建新的 deployment。
	Create(ctx context.Context, d *Deployment) error

	// Update 更新已有的 deployment。
	Update(ctx context.Context, d *Deployment) error

	// Delete 删除 deployment。
	Delete(ctx context.Context, id int64) error
}
