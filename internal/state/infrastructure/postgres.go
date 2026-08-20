package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infra-orchestration/controlplane/internal/state/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) UpsertDesired(ctx context.Context, update domain.DesiredUpdate) (domain.ResourceState, error) {
	var s domain.ResourceState
	err := r.db.QueryRow(ctx, `
		INSERT INTO resource_states(id,resource_id,environment_id,desired_state,actual_state,last_execution_status,version,lock_version,updated_at)
		SELECT $1,r.id,r.environment_id,$2::jsonb,COALESCE(rs.actual_state,'{}'::jsonb),COALESCE(rs.last_execution_status,'pending'),COALESCE(rs.version,0)+1,COALESCE(rs.lock_version,0)+1,$3
		FROM resources r LEFT JOIN resource_states rs ON rs.resource_id=r.id AND rs.version=(SELECT MAX(version) FROM resource_states x WHERE x.resource_id=r.id)
		WHERE r.id=$4
		ON CONFLICT(resource_id,version) DO UPDATE SET desired_state=EXCLUDED.desired_state,updated_at=EXCLUDED.updated_at
		RETURNING id,resource_id,environment_id,desired_state,actual_state,last_execution_status,version,lock_version,updated_at`,
		uuid.NewString(), []byte(update.DesiredState), time.Now().UTC(), update.ResourceID).
		Scan(&s.ID, &s.ResourceID, &s.EnvironmentID, &s.DesiredState, &s.ActualState, &s.LastExecutionStatus, &s.Version, &s.LockVersion, &s.UpdatedAt)
	if err != nil {
		return domain.ResourceState{}, fmt.Errorf("upsert desired state: %w", err)
	}
	return s, nil
}

func (r *PostgresRepository) UpdateActual(ctx context.Context, update domain.ActualUpdate) (domain.ResourceState, error) {
	var s domain.ResourceState
	err := r.db.QueryRow(ctx, `
		UPDATE resource_states SET actual_state=$1::jsonb,last_execution_status=$2,updated_at=$3
		WHERE resource_id=$4 AND version=(SELECT MAX(version) FROM resource_states x WHERE x.resource_id=$4)
		RETURNING id,resource_id,environment_id,desired_state,actual_state,last_execution_status,version,lock_version,updated_at`,
		[]byte(update.ActualState), string(update.Status), time.Now().UTC(), update.ResourceID).
		Scan(&s.ID, &s.ResourceID, &s.EnvironmentID, &s.DesiredState, &s.ActualState, &s.LastExecutionStatus, &s.Version, &s.LockVersion, &s.UpdatedAt)
	if err != nil {
		return domain.ResourceState{}, fmt.Errorf("update actual state: %w", err)
	}
	return s, nil
}

func (r *PostgresRepository) GetByResource(ctx context.Context, resourceID string) (domain.ResourceState, error) {
	var s domain.ResourceState
	err := r.db.QueryRow(ctx, `SELECT id,resource_id,environment_id,desired_state,actual_state,last_execution_status,version,lock_version,updated_at
		FROM resource_states WHERE resource_id=$1 ORDER BY version DESC LIMIT 1`, resourceID).
		Scan(&s.ID, &s.ResourceID, &s.EnvironmentID, &s.DesiredState, &s.ActualState, &s.LastExecutionStatus, &s.Version, &s.LockVersion, &s.UpdatedAt)
	if err != nil {
		return domain.ResourceState{}, fmt.Errorf("get state: %w", err)
	}
	return s, nil
}

func (r *PostgresRepository) ListByEnvironment(ctx context.Context, environmentID string) ([]domain.ResourceState, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (rs.resource_id) rs.id,rs.resource_id,rs.environment_id,rs.desired_state,rs.actual_state,rs.last_execution_status,rs.version,rs.lock_version,rs.updated_at
		FROM resource_states rs WHERE rs.environment_id=$1 ORDER BY rs.resource_id,rs.version DESC`, environmentID)
	if err != nil {
		return nil, fmt.Errorf("list states: %w", err)
	}
	defer rows.Close()
	var out []domain.ResourceState
	for rows.Next() {
		var s domain.ResourceState
		if err := rows.Scan(&s.ID, &s.ResourceID, &s.EnvironmentID, &s.DesiredState, &s.ActualState, &s.LastExecutionStatus, &s.Version, &s.LockVersion, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan state: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) Snapshot(ctx context.Context, resourceID string) (domain.StateSnapshot, error) {
	var s domain.ResourceState
	err := r.db.QueryRow(ctx, `SELECT id,resource_id,environment_id,desired_state,actual_state,last_execution_status,version,lock_version,updated_at
		FROM resource_states WHERE resource_id=$1 ORDER BY version DESC LIMIT 1`, resourceID).
		Scan(&s.ID, &s.ResourceID, &s.EnvironmentID, &s.DesiredState, &s.ActualState, &s.LastExecutionStatus, &s.Version, &s.LockVersion, &s.UpdatedAt)
	if err != nil {
		return domain.StateSnapshot{}, fmt.Errorf("load state for snapshot: %w", err)
	}
	snap := domain.StateSnapshot{
		ID:           uuid.NewString(),
		ResourceID:   resourceID,
		DesiredState: s.DesiredState,
		ActualState:  s.ActualState,
		Status:       s.LastExecutionStatus,
		Version:      s.Version,
		CapturedAt:   time.Now().UTC(),
	}
	_, err = r.db.Exec(ctx, `INSERT INTO state_snapshots(id,resource_id,desired_state,actual_state,status,version,captured_at) VALUES($1,$2,$3,$4,$5,$6,$7)`,
		snap.ID, snap.ResourceID, []byte(snap.DesiredState), []byte(snap.ActualState), string(snap.Status), snap.Version, snap.CapturedAt)
	if err != nil {
		return domain.StateSnapshot{}, fmt.Errorf("insert state snapshot: %w", err)
	}
	return snap, nil
}

func (r *PostgresRepository) ListSnapshots(ctx context.Context, resourceID string, limit int) ([]domain.StateSnapshot, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.Query(ctx, `SELECT id,resource_id,desired_state,actual_state,status,version,captured_at FROM state_snapshots WHERE resource_id=$1 ORDER BY captured_at DESC LIMIT $2`, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer rows.Close()
	var out []domain.StateSnapshot
	for rows.Next() {
		var s domain.StateSnapshot
		if err := rows.Scan(&s.ID, &s.ResourceID, &s.DesiredState, &s.ActualState, &s.Status, &s.Version, &s.CapturedAt); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) Rollback(ctx context.Context, resourceID string, snapshotID string) (domain.ResourceState, error) {
	var snap domain.StateSnapshot
	err := r.db.QueryRow(ctx, `SELECT id,resource_id,desired_state,actual_state,status,version,captured_at FROM state_snapshots WHERE id=$1 AND resource_id=$2`, snapshotID, resourceID).
		Scan(&snap.ID, &snap.ResourceID, &snap.DesiredState, &snap.ActualState, &snap.Status, &snap.Version, &snap.CapturedAt)
	if err != nil {
		return domain.ResourceState{}, fmt.Errorf("get snapshot: %w", err)
	}
	var s domain.ResourceState
	err = r.db.QueryRow(ctx, `
		INSERT INTO resource_states(id,resource_id,environment_id,desired_state,actual_state,last_execution_status,version,lock_version,updated_at)
		SELECT $1,r.id,r.environment_id,$2::jsonb,$3::jsonb,$4,$5,COALESCE(rs.lock_version,0)+1,$6
		FROM resources r LEFT JOIN resource_states rs ON rs.resource_id=r.id AND rs.version=(SELECT MAX(version) FROM resource_states x WHERE x.resource_id=r.id)
		WHERE r.id=$7
		ON CONFLICT(resource_id,version) DO UPDATE SET desired_state=EXCLUDED.desired_state,actual_state=EXCLUDED.actual_state,last_execution_status=EXCLUDED.last_execution_status,updated_at=EXCLUDED.updated_at
		RETURNING id,resource_id,environment_id,desired_state,actual_state,last_execution_status,version,lock_version,updated_at`,
		uuid.NewString(), []byte(snap.DesiredState), []byte(snap.ActualState), string(snap.Status), snap.Version+1, time.Now().UTC(), resourceID).
		Scan(&s.ID, &s.ResourceID, &s.EnvironmentID, &s.DesiredState, &s.ActualState, &s.LastExecutionStatus, &s.Version, &s.LockVersion, &s.UpdatedAt)
	if err != nil {
		return domain.ResourceState{}, fmt.Errorf("rollback state: %w", err)
	}
	return s, nil
}

var _ = pgx.ErrNoRows
