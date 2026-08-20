package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infra-orchestration/controlplane/internal/plan/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, input domain.CreateInput) (domain.Plan, error) {
	if input.ID == "" {
		input.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	var p domain.Plan
	err := r.db.QueryRow(ctx, `
		INSERT INTO plans(id,environment_id,status,diff,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING id,environment_id,status,diff,created_by,created_at,updated_at`,
		input.ID, input.EnvironmentID, string(input.Status), []byte(input.Diff), input.CreatedBy, now, now).
		Scan(&p.ID, &p.EnvironmentID, &p.Status, &p.Diff, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Plan{}, fmt.Errorf("insert plan: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (domain.Plan, error) {
	var p domain.Plan
	err := r.db.QueryRow(ctx, `SELECT id,environment_id,status,diff,created_by,created_at,updated_at FROM plans WHERE id=$1`, id).
		Scan(&p.ID, &p.EnvironmentID, &p.Status, &p.Diff, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Plan{}, fmt.Errorf("get plan: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) ListByEnvironment(ctx context.Context, environmentID string, limit int) ([]domain.Plan, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(ctx, `SELECT id,environment_id,status,diff,created_by,created_at,updated_at FROM plans WHERE environment_id=$1 ORDER BY created_at DESC LIMIT $2`, environmentID, limit)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()
	var out []domain.Plan
	for rows.Next() {
		var p domain.Plan
		if err := rows.Scan(&p.ID, &p.EnvironmentID, &p.Status, &p.Diff, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) (domain.Plan, error) {
	var p domain.Plan
	err := r.db.QueryRow(ctx, `UPDATE plans SET status=$1,updated_at=$2 WHERE id=$3 RETURNING id,environment_id,status,diff,created_by,created_at,updated_at`,
		string(status), time.Now().UTC(), id).
		Scan(&p.ID, &p.EnvironmentID, &p.Status, &p.Diff, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Plan{}, fmt.Errorf("update plan status: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) UpdateDiff(ctx context.Context, id string, diff json.RawMessage) (domain.Plan, error) {
	var p domain.Plan
	err := r.db.QueryRow(ctx, `UPDATE plans SET diff=$1::jsonb,updated_at=$2 WHERE id=$3 RETURNING id,environment_id,status,diff,created_by,created_at,updated_at`,
		[]byte(diff), time.Now().UTC(), id).
		Scan(&p.ID, &p.EnvironmentID, &p.Status, &p.Diff, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Plan{}, fmt.Errorf("update plan diff: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) SaveDiffItems(ctx context.Context, planID string, items []domain.DiffItem) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin diff items: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM plan_diff_items WHERE plan_id=$1`, planID); err != nil {
		return fmt.Errorf("delete old diff items: %w", err)
	}
	for i, item := range items {
		_, err := tx.Exec(ctx, `INSERT INTO plan_diff_items(id,plan_id,position,resource_id,name,type,operation,before,after)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, uuid.NewString(), planID, i, item.ResourceID, item.Name, item.Type, string(item.Operation), []byte(item.Before), []byte(item.After))
		if err != nil {
			return fmt.Errorf("insert diff item: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetDiffItems(ctx context.Context, planID string) ([]domain.DiffItem, error) {
	rows, err := r.db.Query(ctx, `SELECT resource_id,name,type,operation,before,after FROM plan_diff_items WHERE plan_id=$1 ORDER BY position`, planID)
	if err != nil {
		return nil, fmt.Errorf("list diff items: %w", err)
	}
	defer rows.Close()
	var out []domain.DiffItem
	for rows.Next() {
		var item domain.DiffItem
		if err := rows.Scan(&item.ResourceID, &item.Name, &item.Type, &item.Operation, &item.Before, &item.After); err != nil {
			return nil, fmt.Errorf("scan diff item: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
