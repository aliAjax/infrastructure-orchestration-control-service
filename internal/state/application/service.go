package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/infra-orchestration/controlplane/internal/state/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ApplyDesired(ctx context.Context, resourceID string, desired json.RawMessage) (domain.ResourceState, error) {
	if !json.Valid(desired) {
		return domain.ResourceState{}, fmt.Errorf("invalid desired state JSON")
	}
	return s.repo.UpsertDesired(ctx, domain.DesiredUpdate{ResourceID: resourceID, DesiredState: desired})
}

func (s *Service) ReportActual(ctx context.Context, resourceID string, actual json.RawMessage, status domain.ExecutionStatus) (domain.ResourceState, error) {
	if !json.Valid(actual) {
		return domain.ResourceState{}, fmt.Errorf("invalid actual state JSON")
	}
	return s.repo.UpdateActual(ctx, domain.ActualUpdate{ResourceID: resourceID, ActualState: actual, Status: status})
}

func (s *Service) Get(ctx context.Context, resourceID string) (domain.ResourceState, error) {
	return s.repo.GetByResource(ctx, resourceID)
}

func (s *Service) List(ctx context.Context, environmentID string) ([]domain.ResourceState, error) {
	return s.repo.ListByEnvironment(ctx, environmentID)
}

func (s *Service) Snapshot(ctx context.Context, resourceID string) (domain.StateSnapshot, error) {
	return s.repo.Snapshot(ctx, resourceID)
}

func (s *Service) Rollback(ctx context.Context, resourceID, snapshotID string) (domain.ResourceState, error) {
	return s.repo.Rollback(ctx, resourceID, snapshotID)
}
