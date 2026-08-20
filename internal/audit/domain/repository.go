package domain

import "context"

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Event, error)
	List(ctx context.Context, entityType, entityID string, limit int) ([]Event, error)
}
