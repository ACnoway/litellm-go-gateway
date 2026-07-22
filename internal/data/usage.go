package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
	_ "modernc.org/sqlite"
)

// UsageRepo 实现 biz.UsageRepo 接口，使用 SQLite 作为存储。
type UsageRepo struct {
	db *sql.DB
}

// NewUsageRepo 创建 SQLite 存储实现。
func NewUsageRepo(db *sql.DB) biz.UsageRepo {
	return &UsageRepo{db: db}
}

// Create 保存一条使用记录。
func (r *UsageRepo) Create(ctx context.Context, log *biz.UsageLog) error {
	query := `
		INSERT INTO usage_logs (
			request_id, provider, model,
			prompt_tokens, completion_tokens, total_tokens,
			success, error_code, duration, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		log.RequestID, log.Provider, log.Model,
		log.PromptTokens, log.CompletionTokens, log.TotalTokens,
		log.Success, log.ErrorCode, log.Duration, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert usage log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	log.ID = id
	return nil
}

// Query 查询使用记录，支持时间范围、provider 过滤和分页。
func (r *UsageRepo) Query(ctx context.Context, opts biz.QueryOptions) ([]*biz.UsageLog, error) {
	var conditions []string
	var args []interface{}

	if opts.StartTime != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *opts.StartTime)
	}
	if opts.EndTime != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *opts.EndTime)
	}
	if opts.Provider != "" {
		conditions = append(conditions, "provider = ?")
		args = append(args, opts.Provider)
	}

	query := "SELECT id, request_id, provider, model, prompt_tokens, completion_tokens, total_tokens, success, error_code, duration, created_at FROM usage_logs"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage logs: %w", err)
	}
	defer rows.Close()

	var logs []*biz.UsageLog
	for rows.Next() {
		log := &biz.UsageLog{}
		err := rows.Scan(
			&log.ID, &log.RequestID, &log.Provider, &log.Model,
			&log.PromptTokens, &log.CompletionTokens, &log.TotalTokens,
			&log.Success, &log.ErrorCode, &log.Duration, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return logs, nil
}

// InitDB 初始化 SQLite 数据库连接并创建表结构。
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 创建表结构
	schema := `
	CREATE TABLE IF NOT EXISTS usage_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request_id TEXT NOT NULL,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		prompt_tokens INTEGER NOT NULL DEFAULT 0,
		completion_tokens INTEGER NOT NULL DEFAULT 0,
		total_tokens INTEGER NOT NULL DEFAULT 0,
		success BOOLEAN NOT NULL DEFAULT 1,
		error_code TEXT,
		duration INTEGER NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_created_at ON usage_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_provider ON usage_logs(provider);
	CREATE INDEX IF NOT EXISTS idx_request_id ON usage_logs(request_id);

	CREATE TABLE IF NOT EXISTS routing_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT NOT NULL UNIQUE,
		providers TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_pattern ON routing_rules(pattern);

	CREATE TABLE IF NOT EXISTS deployments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		actual_model TEXT NOT NULL,
		providers TEXT NOT NULL,
		strategy TEXT NOT NULL DEFAULT 'priority',
		weights TEXT,
		max_tokens INTEGER DEFAULT 0,
		description TEXT,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_deployment_name ON deployments(name);
	CREATE INDEX IF NOT EXISTS idx_deployment_enabled ON deployments(enabled);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}
