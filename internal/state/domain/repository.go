package domain

import "context"

type Repository interface {
	UpsertDesired(ctx context.Context, update DesiredUpdate) (ResourceState, error)
	UpdateActual(ctx context.Context, update ActualUpdate) (ResourceState, error)
	GetByResource(ctx context.Context, resourceID string) (ResourceState, error)
	ListByEnvironment(ctx context.Context, environmentID string) ([]ResourceState, error)
	Snapshot(ctx context.Context, resourceID string) (StateSnapshot, error)
	ListSnapshots(ctx context.Context, resourceID string, limit int) ([]StateSnapshot, error)
	Rollback(ctx context.Context, resourceID string, snapshotID string) (ResourceState, error)
}
