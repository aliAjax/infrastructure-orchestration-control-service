package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infra-orchestration/controlplane/internal/execution/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTask(ctx context.Context, input domain.CreateTaskInput) (domain.Task, error) {
	tasks, err := r.CreateTasks(ctx, []domain.CreateTaskInput{input})
	if err != nil {
		return domain.Task{}, err
	}
	return tasks[0], nil
}

func (r *PostgresRepository) CreateTasks(ctx context.Context, inputs []domain.CreateTaskInput) ([]domain.Task, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create tasks: %w", err)
	}
	defer tx.Rollback(ctx)
	out := make([]domain.Task, 0, len(inputs))
	for _, input := range inputs {
		var t domain.Task
		err := tx.QueryRow(ctx, `
			INSERT INTO execution_tasks(id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
			RETURNING id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at`,
			input.ID, input.PlanID, input.ResourceID, input.EnvironmentID, string(domain.StatusPending), 0, input.MaxAttempts, []byte(input.Input), []byte(`{}`), "", input.LockKey, input.Timeout, "", time.Now().UTC(), nil, nil, time.Now().UTC()).
			Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt)
		if err != nil {
			return nil, fmt.Errorf("insert execution task: %w", err)
		}
		out = append(out, t)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tasks: %w", err)
	}
	return out, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (domain.Task, error) {
	var t domain.Task
	err := r.db.QueryRow(ctx, `SELECT id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at FROM execution_tasks WHERE id=$1`, id).
		Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) ListByPlan(ctx context.Context, planID string) ([]domain.Task, error) {
	rows, err := r.db.Query(ctx, `SELECT id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at FROM execution_tasks WHERE plan_id=$1 ORDER BY created_at`, planID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()
	var out []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) ClaimNext(ctx context.Context, owner string) (*domain.Task, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin claim: %w", err)
	}
	defer tx.Rollback(ctx)
	var id string
	err = tx.QueryRow(ctx, `
		SELECT id FROM execution_tasks
		WHERE status IN ('pending','failed') AND next_run_at<=$1 AND (lease_owner='' OR lease_owner IS NULL)
		ORDER BY next_run_at,created_at
		FOR UPDATE SKIP LOCKED LIMIT 1`, time.Now().UTC()).Scan(&id)
	if err == pgx.ErrNoRows {
		_ = tx.Rollback(ctx)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find claimable task: %w", err)
	}
	now := time.Now().UTC()
	var t domain.Task
	err = tx.QueryRow(ctx, `
		UPDATE execution_tasks SET status=$1,attempt=attempt+1,lease_owner=$2,started_at=$3,next_run_at=$4,error=''
		WHERE id=$5
		RETURNING id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at`,
		string(domain.StatusClaimed), owner, now, now.Add(5*time.Minute), id).
		Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt)
	if err != nil {
		return nil, fmt.Errorf("claim task: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit claim: %w", err)
	}
	return &t, nil
}

func (r *PostgresRepository) Complete(ctx context.Context, id, owner string, output json.RawMessage) (domain.Task, error) {
	now := time.Now().UTC()
	var t domain.Task
	err := r.db.QueryRow(ctx, `
		UPDATE execution_tasks SET status=$1,output=$2::jsonb,error='',completed_at=$3,lease_owner=''
		WHERE id=$4 AND lease_owner=$5
		RETURNING id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at`,
		string(domain.StatusSucceeded), []byte(output), now, id, owner).
		Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt)
	if err != nil {
		return domain.Task{}, fmt.Errorf("complete task: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) Fail(ctx context.Context, id, owner, message string) (domain.Task, error) {
	now := time.Now().UTC()
	var t domain.Task
	err := r.db.QueryRow(ctx, `
		UPDATE execution_tasks SET status=$1,error=$2,completed_at=$3,lease_owner='',next_run_at=$4
		WHERE id=$5 AND lease_owner=$6
		RETURNING id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at`,
		string(domain.StatusFailed), message, now, now.Add(time.Minute), id, owner).
		Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt)
	if err != nil {
		return domain.Task{}, fmt.Errorf("fail task: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) Retry(ctx context.Context, id string) (domain.Task, error) {
	var t domain.Task
	err := r.db.QueryRow(ctx, `UPDATE execution_tasks SET status='pending',next_run_at=$1,error='',lease_owner='' WHERE id=$2 RETURNING id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at`,
		time.Now().UTC(), id).
		Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt)
	if err != nil {
		return domain.Task{}, fmt.Errorf("retry task: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) Cancel(ctx context.Context, id string) (domain.Task, error) {
	var t domain.Task
	err := r.db.QueryRow(ctx, `UPDATE execution_tasks SET status='cancelled',lease_owner='' WHERE id=$1 RETURNING id,plan_id,resource_id,environment_id,status,attempt,max_attempts,input,output,error,lock_key,timeout,lease_owner,created_at,started_at,completed_at,next_run_at`, id).
		Scan(&t.ID, &t.PlanID, &t.ResourceID, &t.EnvironmentID, &t.Status, &t.Attempt, &t.MaxAttempts, &t.Input, &t.Output, &t.Error, &t.LockKey, &t.Timeout, &t.LeaseOwner, &t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.NextRunAt)
	if err != nil {
		return domain.Task{}, fmt.Errorf("cancel task: %w", err)
	}
	return t, nil
}
