package domain

import (
	"context"
	"encoding/json"
)

type Repository interface {
	CreateTask(ctx context.Context, input CreateTaskInput) (Task, error)
	CreateTasks(ctx context.Context, inputs []CreateTaskInput) ([]Task, error)
	Get(ctx context.Context, id string) (Task, error)
	ListByPlan(ctx context.Context, planID string) ([]Task, error)
	ClaimNext(ctx context.Context, owner string) (*Task, error)
	Complete(ctx context.Context, id, owner string, output json.RawMessage) (Task, error)
	Fail(ctx context.Context, id, owner, message string) (Task, error)
	Retry(ctx context.Context, id string) (Task, error)
	Cancel(ctx context.Context, id string) (Task, error)
}
