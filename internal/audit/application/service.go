package application

import (
	"context"
	"fmt"

	"github.com/infra-orchestration/controlplane/internal/audit/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, input domain.CreateInput) (domain.Event, error) {
	if input.Type == "" || input.Actor == "" {
		return domain.Event{}, fmt.Errorf("type and actor are required")
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) List(ctx context.Context, entityType, entityID string, limit int) ([]domain.Event, error) {
	return s.repo.List(ctx, entityType, entityID, limit)
}
