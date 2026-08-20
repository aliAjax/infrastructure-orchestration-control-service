package application

import (
	"context"
	"fmt"

	graphdomain "github.com/infra-orchestration/controlplane/internal/graph/domain"
	"github.com/infra-orchestration/controlplane/internal/resource/domain"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Build(resources []domain.Resource) *graphdomain.Graph {
	g := graphdomain.NewGraph()
	ids := make(map[string]struct{}, len(resources))
	for _, r := range resources {
		ids[r.ID] = struct{}{}
	}
	for _, r := range resources {
		deps := append([]string(nil), r.DependsOn...)
		g.AddNode(graphdomain.Node{ID: r.ID, Name: r.Name, Dependencies: deps})
	}
	return g
}

func (s *Service) ValidateDependencies(ctx context.Context, resources []domain.Resource) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	g := s.Build(resources)
	if err := g.Validate(); err != nil {
		return fmt.Errorf("graph validation: %w", err)
	}
	return nil
}

func (s *Service) Levels(ctx context.Context, resources []domain.Resource) ([][]string, error) {
	if err := s.ValidateDependencies(ctx, resources); err != nil {
		return nil, err
	}
	return s.Build(resources).TopologicalLevels()
}
