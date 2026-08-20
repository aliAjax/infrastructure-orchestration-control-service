package application

import (
	"context"
	"fmt"

	"github.com/infra-orchestration/controlplane/internal/drift/domain"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
)

func (s *Service) ReconcileBatch(ctx context.Context, environmentID string) ([]domain.DriftRecord, error) {
	records, err := s.repo.ListByEnvironment(ctx, environmentID, true)
	if err != nil {
		return nil, fmt.Errorf("list unresolved drift: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	plan, err := s.planService.Generate(ctx, plandomain.GenerateInput{EnvironmentID: environmentID, CreatedBy: "drift-detector"})
	if err != nil {
		return nil, fmt.Errorf("generate remedy plan: %w", err)
	}
	for _, record := range records {
		if _, err := s.repo.Resolve(ctx, domain.ResolveInput{ID: record.ID, RemedyPlanID: plan.ID}); err != nil {
			return nil, fmt.Errorf("resolve drift record: %w", err)
		}
	}
	return records, nil
}
