package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/acnoway/litellm-go-gateway/internal/biz"
)

// DeploymentRepo 实现 biz.DeploymentRepo 接口，使用 SQLite 作为存储。
type DeploymentRepo struct {
	db *sql.DB
}

// NewDeploymentRepo 创建 SQLite 存储实现。
func NewDeploymentRepo(db *sql.DB) biz.DeploymentRepo {
	return &DeploymentRepo{db: db}
}

// List 返回所有 deployments。
func (r *DeploymentRepo) List(ctx context.Context) ([]*biz.Deployment, error) {
	query := `
		SELECT id, name, actual_model, providers, strategy, weights, max_tokens, description, enabled, created_at, updated_at
		FROM deployments
		ORDER BY created_at DESC
	`
	return r.query(ctx, query)
}

// ListEnabled 返回所有启用的 deployments。
func (r *DeploymentRepo) ListEnabled(ctx context.Context) ([]*biz.Deployment, error) {
	query := `
		SELECT id, name, actual_model, providers, strategy, weights, max_tokens, description, enabled, created_at, updated_at
		FROM deployments
		WHERE enabled = 1
		ORDER BY created_at DESC
	`
	return r.query(ctx, query)
}

// Get 根据 ID 获取单个 deployment。
func (r *DeploymentRepo) Get(ctx context.Context, id int64) (*biz.Deployment, error) {
	query := `
		SELECT id, name, actual_model, providers, strategy, weights, max_tokens, description, enabled, created_at, updated_at
		FROM deployments
		WHERE id = ?
	`
	deployments, err := r.query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	if len(deployments) == 0 {
		return nil, fmt.Errorf("deployment not found: id=%d", id)
	}
	return deployments[0], nil
}

// GetByName 根据逻辑模型名获取 deployment。
func (r *DeploymentRepo) GetByName(ctx context.Context, name string) (*biz.Deployment, error) {
	query := `
		SELECT id, name, actual_model, providers, strategy, weights, max_tokens, description, enabled, created_at, updated_at
		FROM deployments
		WHERE name = ?
	`
	deployments, err := r.query(ctx, query, name)
	if err != nil {
		return nil, err
	}
	if len(deployments) == 0 {
		return nil, fmt.Errorf("deployment not found: name=%s", name)
	}
	return deployments[0], nil
}

// Create 创建新的 deployment。
func (r *DeploymentRepo) Create(ctx context.Context, d *biz.Deployment) error {
	providersJSON, err := json.Marshal(d.Providers)
	if err != nil {
		return fmt.Errorf("failed to marshal providers: %w", err)
	}

	weightsJSON, err := json.Marshal(d.Weights)
	if err != nil {
		return fmt.Errorf("failed to marshal weights: %w", err)
	}

	query := `
		INSERT INTO deployments (
			name, actual_model, providers, strategy, weights, max_tokens, description, enabled, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		d.Name, d.ActualModel, string(providersJSON), d.Strategy, string(weightsJSON),
		d.MaxTokens, d.Description, d.Enabled, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert deployment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	d.ID = id
	d.CreatedAt = now
	d.UpdatedAt = now
	return nil
}

// Update 更新已有的 deployment。
func (r *DeploymentRepo) Update(ctx context.Context, d *biz.Deployment) error {
	providersJSON, err := json.Marshal(d.Providers)
	if err != nil {
		return fmt.Errorf("failed to marshal providers: %w", err)
	}

	weightsJSON, err := json.Marshal(d.Weights)
	if err != nil {
		return fmt.Errorf("failed to marshal weights: %w", err)
	}

	query := `
		UPDATE deployments
		SET name = ?, actual_model = ?, providers = ?, strategy = ?, weights = ?,
		    max_tokens = ?, description = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		d.Name, d.ActualModel, string(providersJSON), d.Strategy, string(weightsJSON),
		d.MaxTokens, d.Description, d.Enabled, now, d.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("deployment not found: id=%d", d.ID)
	}

	d.UpdatedAt = now
	return nil
}

// Delete 删除 deployment。
func (r *DeploymentRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM deployments WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("deployment not found: id=%d", id)
	}

	return nil
}

// query 是通用的查询辅助函数，处理 deployment 的扫描和 JSON 反序列化。
func (r *DeploymentRepo) query(ctx context.Context, query string, args ...interface{}) ([]*biz.Deployment, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*biz.Deployment
	for rows.Next() {
		d := &biz.Deployment{}
		var providersJSON, weightsJSON string

		err := rows.Scan(
			&d.ID, &d.Name, &d.ActualModel, &providersJSON, &d.Strategy, &weightsJSON,
			&d.MaxTokens, &d.Description, &d.Enabled, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}

		// 反序列化 JSON 字段
		if err := json.Unmarshal([]byte(providersJSON), &d.Providers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal providers: %w", err)
		}
		if weightsJSON != "" && weightsJSON != "null" {
			if err := json.Unmarshal([]byte(weightsJSON), &d.Weights); err != nil {
				return nil, fmt.Errorf("failed to unmarshal weights: %w", err)
			}
		}

		deployments = append(deployments, d)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return deployments, nil
}
