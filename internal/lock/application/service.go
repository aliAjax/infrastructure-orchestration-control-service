package application

import (
	"context"
	"fmt"
	"time"

	"github.com/infra-orchestration/controlplane/internal/lock/domain"
)

type Service struct {
	manager domain.Manager
}

func NewService(manager domain.Manager) *Service {
	return &Service{manager: manager}
}

func (s *Service) WithLock(ctx context.Context, key, owner string, ttl time.Duration, fn func(ctx context.Context) error) error {
	lock, err := s.manager.Acquire(ctx, key, owner, ttl)
	if err != nil {
		return fmt.Errorf("acquire lock %q: %w", key, err)
	}
	defer s.manager.Release(context.Background(), lock)
	renewCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go s.renew(renewCtx, lock, ttl)
	if err := fn(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) renew(ctx context.Context, lock *domain.Lock, ttl time.Duration) {
	ticker := time.NewTicker(ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.manager.Renew(ctx, lock, ttl); err != nil {
				return
			}
		}
	}
}
