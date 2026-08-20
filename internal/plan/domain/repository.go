package domain

import (
	"context"
	"encoding/json"
)

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Plan, error)
	Get(ctx context.Context, id string) (Plan, error)
	ListByEnvironment(ctx context.Context, environmentID string, limit int) ([]Plan, error)
	UpdateStatus(ctx context.Context, id string, status Status) (Plan, error)
	UpdateDiff(ctx context.Context, id string, diff json.RawMessage) (Plan, error)
	SaveDiffItems(ctx context.Context, planID string, items []DiffItem) error
	GetDiffItems(ctx context.Context, planID string) ([]DiffItem, error)
}
