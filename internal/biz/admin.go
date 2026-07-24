package biz

// ProviderInfo 表示一个 provider 及其状态
type ProviderInfo struct {
	Name      string `json:"name"`       // provider 名称
	Available bool   `json:"available"`  // 是否可用（配置了 API Key）
	Type      string `json:"type"`       // provider 类型
	BaseURL   string `json:"base_url"`   // base URL
}

// RoutingRuleRequest 用于创建或更新路由规则
type RoutingRuleRequest struct {
	Pattern   string   `json:"pattern" binding:"required"`   // 模型名匹配模式
	Providers []string `json:"providers" binding:"required"` // provider 列表（按优先级）
}

// RoutingRuleResponse 表示一条路由规则
type RoutingRuleResponse struct {
	ID        int      `json:"id"`        // 规则 ID
	Pattern   string   `json:"pattern"`   // 模型名匹配模式
	Providers []string `json:"providers"` // provider 列表
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// RoutingRuleRepo 定义路由规则持久化接口
type RoutingRuleRepo interface {
	List() ([]RoutingRuleResponse, error)
	Get(id int) (*RoutingRuleResponse, error)
	Create(pattern string, providers []string) (*RoutingRuleResponse, error)
	Update(id int, pattern string, providers []string) (*RoutingRuleResponse, error)
	Delete(id int) error
}

// ModelInfo 表示一个可用的逻辑模型（用于 /v1/models 端点）
type ModelInfo struct {
	ID          string   `json:"id"`           // 逻辑模型名（用户可见）
	Object      string   `json:"object"`       // 固定为 "model"
	Created     int64    `json:"created"`      // 时间戳
	OwnedBy     string   `json:"owned_by"`     // provider 名称
	Ready       bool     `json:"ready"`        // 是否就绪（enabled）
	Description string   `json:"description"`  // 描述
	Providers   []string `json:"providers"`    // 可用的 providers
}
