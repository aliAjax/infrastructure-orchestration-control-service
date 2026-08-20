package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/infra-orchestration/controlplane/internal/resource/domain"
)

type ExecutorFunc func(ctx context.Context, resourceID string) error

func (s *Service) ExecuteLevels(ctx context.Context, resources []domain.Resource, fn ExecutorFunc) error {
	levels, err := s.Levels(ctx, resources)
	if err != nil {
		return err
	}
	byID := make(map[string]domain.Resource, len(resources))
	for _, resource := range resources {
		byID[resource.ID] = resource
	}
	for _, level := range levels {
		var wg sync.WaitGroup
		errCh := make(chan error, len(level))
		for _, id := range level {
			if _, ok := byID[id]; !ok {
				continue
			}
			wg.Add(1)
			go func(resourceID string) {
				defer wg.Done()
				if err := fn(ctx, resourceID); err != nil {
					errCh <- err
				}
			}(id)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				return fmt.Errorf("execute graph level: %w", err)
			}
		}
	}
	return nil
}
