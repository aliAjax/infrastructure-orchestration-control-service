package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infra-orchestration/controlplane/internal/approval/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, input domain.CreateInput) (domain.Request, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	var request domain.Request
	err := r.db.QueryRow(ctx, `
		INSERT INTO approval_requests(id,plan_id,environment_id,requested_by,reason,status,approved_by,decision_note,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id,plan_id,environment_id,requested_by,reason,status,approved_by,decision_note,created_at,updated_at`,
		id, input.PlanID, input.EnvironmentID, input.RequestedBy, input.Reason, string(domain.StatusPending), "", "", now, now).
		Scan(&request.ID, &request.PlanID, &request.EnvironmentID, &request.RequestedBy, &request.Reason, &request.Status, &request.ApprovedBy, &request.DecisionNote, &request.CreatedAt, &request.UpdatedAt)
	if err != nil {
		return domain.Request{}, fmt.Errorf("insert approval request: %w", err)
	}
	return request, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (domain.Request, error) {
	var request domain.Request
	err := r.db.QueryRow(ctx, `SELECT id,plan_id,environment_id,requested_by,reason,status,approved_by,decision_note,created_at,updated_at FROM approval_requests WHERE id=$1`, id).
		Scan(&request.ID, &request.PlanID, &request.EnvironmentID, &request.RequestedBy, &request.Reason, &request.Status, &request.ApprovedBy, &request.DecisionNote, &request.CreatedAt, &request.UpdatedAt)
	if err != nil {
		return domain.Request{}, fmt.Errorf("get approval request: %w", err)
	}
	return request, nil
}

func (r *PostgresRepository) ListByPlan(ctx context.Context, planID string) ([]domain.Request, error) {
	rows, err := r.db.Query(ctx, `SELECT id,plan_id,environment_id,requested_by,reason,status,approved_by,decision_note,created_at,updated_at FROM approval_requests WHERE plan_id=$1 ORDER BY created_at`, planID)
	if err != nil {
		return nil, fmt.Errorf("list approval requests: %w", err)
	}
	defer rows.Close()
	var out []domain.Request
	for rows.Next() {
		var request domain.Request
		if err := rows.Scan(&request.ID, &request.PlanID, &request.EnvironmentID, &request.RequestedBy, &request.Reason, &request.Status, &request.ApprovedBy, &request.DecisionNote, &request.CreatedAt, &request.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan approval request: %w", err)
		}
		out = append(out, request)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) Decide(ctx context.Context, input domain.DecisionInput) (domain.Request, error) {
	status := domain.StatusRejected
	if input.Approved {
		status = domain.StatusApproved
	}
	var request domain.Request
	err := r.db.QueryRow(ctx, `
		UPDATE approval_requests SET status=$1,approved_by=$2,decision_note=$3,updated_at=$4 WHERE id=$5
		RETURNING id,plan_id,environment_id,requested_by,reason,status,approved_by,decision_note,created_at,updated_at`,
		string(status), input.ApprovedBy, input.DecisionNote, time.Now().UTC(), input.ID).
		Scan(&request.ID, &request.PlanID, &request.EnvironmentID, &request.RequestedBy, &request.Reason, &request.Status, &request.ApprovedBy, &request.DecisionNote, &request.CreatedAt, &request.UpdatedAt)
	if err != nil {
		return domain.Request{}, fmt.Errorf("decide approval request: %w", err)
	}
	return request, nil
}
