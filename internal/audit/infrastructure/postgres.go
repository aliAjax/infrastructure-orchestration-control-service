package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infra-orchestration/controlplane/internal/audit/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, input domain.CreateInput) (domain.Event, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	var e domain.Event
	err := r.db.QueryRow(ctx, `
		INSERT INTO audit_events(id,type,actor,entity_type,entity_id,content,result,error,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id,type,actor,entity_type,entity_id,content,result,error,created_at`,
		id, string(input.Type), input.Actor, input.EntityType, input.EntityID, []byte(input.Content), input.Result, input.Error, now).
		Scan(&e.ID, &e.Type, &e.Actor, &e.EntityType, &e.EntityID, &e.Content, &e.Result, &e.Error, &e.CreatedAt)
	if err != nil {
		return domain.Event{}, fmt.Errorf("insert audit event: %w", err)
	}
	return e, nil
}

func (r *PostgresRepository) List(ctx context.Context, entityType, entityID string, limit int) ([]domain.Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `SELECT id,type,actor,entity_type,entity_id,content,result,error,created_at FROM audit_events`
	args := []any{}
	if entityType != "" {
		query += ` WHERE entity_type=$1`
		args = append(args, entityType)
		if entityID != "" {
			query += ` AND entity_id=$2`
			args = append(args, entityID)
		}
	} else if entityID != "" {
		query += ` WHERE entity_id=$1`
		args = append(args, entityID)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, limit)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()
	var out []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.Type, &e.Actor, &e.EntityType, &e.EntityID, &e.Content, &e.Result, &e.Error, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
