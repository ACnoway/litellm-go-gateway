package biz

import (
	"context"
	"time"
)

// UsageLog 是单次请求的 token 消耗记录。
type UsageLog struct {
	ID               int64     `json:"id"`
	RequestID        string    `json:"request_id"`        // 请求唯一标识
	Provider         string    `json:"provider"`          // 使用的 provider（openai/anthropic/azure）
	Model            string    `json:"model"`             // 请求的模型名称
	PromptTokens     int       `json:"prompt_tokens"`     // 输入 token 数
	CompletionTokens int       `json:"completion_tokens"` // 输出 token 数
	TotalTokens      int       `json:"total_tokens"`      // 总 token 数
	Success          bool      `json:"success"`           // 请求是否成功
	ErrorCode        string    `json:"error_code"`        // 失败时的错误码
	Duration         int64     `json:"duration"`          // 请求耗时（毫秒）
	CreatedAt        time.Time `json:"created_at"`        // 记录创建时间
}

// UsageRepo 定义使用日志的存储接口。
// 数据层（internal/data）负责实现此接口。
type UsageRepo interface {
	// Create 保存一条使用记录
	Create(ctx context.Context, log *UsageLog) error

	// Query 查询使用记录，支持按时间范围和 provider 过滤
	Query(ctx context.Context, opts QueryOptions) ([]*UsageLog, error)
}

// QueryOptions 是查询使用日志的过滤条件。
type QueryOptions struct {
	StartTime *time.Time
	EndTime   *time.Time
	Provider  string
	Limit     int
	Offset    int
}
