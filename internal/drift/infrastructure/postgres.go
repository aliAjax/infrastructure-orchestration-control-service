package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infra-orchestration/controlplane/internal/drift/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, input domain.CreateInput) (domain.DriftRecord, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	var d domain.DriftRecord
	err := r.db.QueryRow(ctx, `
		INSERT INTO drift_records(id,environment_id,resource_id,desired_state,actual_state,severity,detected_at,resolved,remedy_plan_id)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id,environment_id,resource_id,desired_state,actual_state,severity,detected_at,resolved,remedy_plan_id`,
		id, input.EnvironmentID, input.ResourceID, []byte(input.DesiredState), []byte(input.ActualState), input.Severity, now, false, input.RemedyPlanID).
		Scan(&d.ID, &d.EnvironmentID, &d.ResourceID, &d.DesiredState, &d.ActualState, &d.Severity, &d.DetectedAt, &d.Resolved, &d.RemedyPlanID)
	if err != nil {
		return domain.DriftRecord{}, fmt.Errorf("insert drift record: %w", err)
	}
	return d, nil
}

func (r *PostgresRepository) ListByEnvironment(ctx context.Context, environmentID string, unresolvedOnly bool) ([]domain.DriftRecord, error) {
	query := `SELECT id,environment_id,resource_id,desired_state,actual_state,severity,detected_at,resolved,remedy_plan_id FROM drift_records WHERE environment_id=$1`
	args := []any{environmentID}
	if unresolvedOnly {
		query += ` AND resolved=false`
	}
	query += ` ORDER BY detected_at DESC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list drift: %w", err)
	}
	defer rows.Close()
	var out []domain.DriftRecord
	for rows.Next() {
		var d domain.DriftRecord
		if err := rows.Scan(&d.ID, &d.EnvironmentID, &d.ResourceID, &d.DesiredState, &d.ActualState, &d.Severity, &d.DetectedAt, &d.Resolved, &d.RemedyPlanID); err != nil {
			return nil, fmt.Errorf("scan drift: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) Resolve(ctx context.Context, input domain.ResolveInput) (domain.DriftRecord, error) {
	var d domain.DriftRecord
	err := r.db.QueryRow(ctx, `UPDATE drift_records SET resolved=true,remedy_plan_id=$1 WHERE id=$2 RETURNING id,environment_id,resource_id,desired_state,actual_state,severity,detected_at,resolved,remedy_plan_id`,
		input.RemedyPlanID, input.ID).
		Scan(&d.ID, &d.EnvironmentID, &d.ResourceID, &d.DesiredState, &d.ActualState, &d.Severity, &d.DetectedAt, &d.Resolved, &d.RemedyPlanID)
	if err != nil {
		return domain.DriftRecord{}, fmt.Errorf("resolve drift: %w", err)
	}
	return d, nil
}
