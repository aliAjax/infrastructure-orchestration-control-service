package application

import (
	"context"
	"fmt"

	graphdomain "github.com/infra-orchestration/controlplane/internal/graph/domain"
	"github.com/infra-orchestration/controlplane/internal/resource/domain"
)

func (s *Service) DependencyPath(ctx context.Context, resources []domain.Resource, from, to string) ([]string, error) {
	if err := s.ValidateDependencies(ctx, resources); err != nil {
		return nil, err
	}
	g := s.Build(resources)
	path := []string{from}
	visited := map[string]bool{}
	var walk func(string) bool
	walk = func(id string) bool {
		if id == to {
			return true
		}
		node, ok := g.Node(id)
		if !ok || visited[id] {
			return false
		}
		visited[id] = true
		for _, dep := range node.Dependencies {
			path = append(path, dep)
			if walk(dep) {
				return true
			}
			path = path[:len(path)-1]
		}
		return false
	}
	if !walk(from) {
		return nil, fmt.Errorf("no dependency path from %s to %s", from, to)
	}
	return path, nil
}

func (s *Service) ReverseDependencies(ctx context.Context, resources []domain.Resource, id string) ([]string, error) {
	if err := s.ValidateDependencies(ctx, resources); err != nil {
		return nil, err
	}
	var out []string
	for _, resource := range resources {
		for _, dep := range resource.DependsOn {
			if dep == id {
				out = append(out, resource.ID)
			}
		}
	}
	return out, nil
}

var _ = graphdomain.NewGraph
