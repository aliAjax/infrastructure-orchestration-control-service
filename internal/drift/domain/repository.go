package domain

import "context"

type Repository interface {
	Create(ctx context.Context, input CreateInput) (DriftRecord, error)
	ListByEnvironment(ctx context.Context, environmentID string, unresolvedOnly bool) ([]DriftRecord, error)
	Resolve(ctx context.Context, input ResolveInput) (DriftRecord, error)
}
