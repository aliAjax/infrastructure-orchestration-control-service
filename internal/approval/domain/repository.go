package domain

import "context"

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Request, error)
	Get(ctx context.Context, id string) (Request, error)
	ListByPlan(ctx context.Context, planID string) ([]Request, error)
	Decide(ctx context.Context, input DecisionInput) (Request, error)
}
