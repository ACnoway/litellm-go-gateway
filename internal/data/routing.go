package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// RoutingRuleRepo 实现路由规则的 SQLite 持久化
type RoutingRuleRepo struct {
	db *sql.DB
}

// NewRoutingRuleRepo 创建路由规则仓储
func NewRoutingRuleRepo(db *sql.DB) *RoutingRuleRepo {
	return &RoutingRuleRepo{db: db}
}

// List 返回所有路由规则
func (r *RoutingRuleRepo) List() ([]biz.RoutingRuleResponse, error) {
	rows, err := r.db.Query(`
		SELECT id, pattern, providers, created_at, updated_at
		FROM routing_rules
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query routing rules: %w", err)
	}
	defer rows.Close()

	var rules []biz.RoutingRuleResponse
	for rows.Next() {
		var rule biz.RoutingRuleResponse
		var providersJSON string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&rule.ID, &rule.Pattern, &providersJSON, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan routing rule: %w", err)
		}

		if err := json.Unmarshal([]byte(providersJSON), &rule.Providers); err != nil {
			return nil, fmt.Errorf("unmarshal providers: %w", err)
		}

		rule.CreatedAt = createdAt.Format(time.RFC3339)
		rule.UpdatedAt = updatedAt.Format(time.RFC3339)
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// Get 返回单条路由规则
func (r *RoutingRuleRepo) Get(id int) (*biz.RoutingRuleResponse, error) {
	var rule biz.RoutingRuleResponse
	var providersJSON string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(`
		SELECT id, pattern, providers, created_at, updated_at
		FROM routing_rules
		WHERE id = ?
	`, id).Scan(&rule.ID, &rule.Pattern, &providersJSON, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("routing rule not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query routing rule: %w", err)
	}

	if err := json.Unmarshal([]byte(providersJSON), &rule.Providers); err != nil {
		return nil, fmt.Errorf("unmarshal providers: %w", err)
	}

	rule.CreatedAt = createdAt.Format(time.RFC3339)
	rule.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &rule, nil
}

// Create 创建新的路由规则
func (r *RoutingRuleRepo) Create(pattern string, providers []string) (*biz.RoutingRuleResponse, error) {
	providersJSON, err := json.Marshal(providers)
	if err != nil {
		return nil, fmt.Errorf("marshal providers: %w", err)
	}

	now := time.Now()
	result, err := r.db.Exec(`
		INSERT INTO routing_rules (pattern, providers, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, pattern, string(providersJSON), now, now)

	if err != nil {
		return nil, fmt.Errorf("insert routing rule: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &biz.RoutingRuleResponse{
		ID:        int(id),
		Pattern:   pattern,
		Providers: providers,
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	}, nil
}

// Update 更新路由规则
func (r *RoutingRuleRepo) Update(id int, pattern string, providers []string) (*biz.RoutingRuleResponse, error) {
	providersJSON, err := json.Marshal(providers)
	if err != nil {
		return nil, fmt.Errorf("marshal providers: %w", err)
	}

	now := time.Now()
	result, err := r.db.Exec(`
		UPDATE routing_rules
		SET pattern = ?, providers = ?, updated_at = ?
		WHERE id = ?
	`, pattern, string(providersJSON), now, id)

	if err != nil {
		return nil, fmt.Errorf("update routing rule: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("routing rule not found")
	}

	return r.Get(id)
}

// Delete 删除路由规则
func (r *RoutingRuleRepo) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM routing_rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete routing rule: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("routing rule not found")
	}

	return nil
}
