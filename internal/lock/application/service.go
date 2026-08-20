package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/infra-orchestration/controlplane/internal/lock/domain"
)

type Service struct {
	manager domain.Manager
}

func NewService(manager domain.Manager) *Service {
	return &Service{manager: manager}
}

func (s *Service) WithLock(ctx context.Context, key, owner string, ttl time.Duration, fn func(ctx context.Context) error) (err error) {
	if ttl <= 0 {
		return fmt.Errorf("acquire lock %q: non-positive ttl %s", key, ttl)
	}
	lock, acquireErr := s.manager.Acquire(ctx, key, owner, ttl)
	if acquireErr != nil {
		return fmt.Errorf("acquire lock %q: %w", key, acquireErr)
	}
	if lock == nil {
		return fmt.Errorf("acquire lock %q: nil lock", key)
	}

	// Release must run even if the caller's context is cancelled, but it still
	// needs the caller's values (trace ids, etc.). WithoutCancel detaches the
	// cancellation while preserving the value chain.
	releaseCtx := context.WithoutCancel(ctx)
	defer func() {
		if releaseErr := s.manager.Release(releaseCtx, lock); releaseErr != nil {
			err = errors.Join(err, fmt.Errorf("release lock %q: %w", key, releaseErr))
		}
	}()

	// Renewal runs on its own context so a renewal failure never cancels the
	// critical section, and so cancellation stops the ticker. Its error is
	// surfaced back to the caller once the goroutine exits.
	renewCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	renewErrCh := make(chan error, 1)
	var renewDone sync.WaitGroup
	renewDone.Add(1)
	go func() {
		defer renewDone.Done()
		renewErrCh <- s.renew(renewCtx, lock, ttl)
	}()

	opErr := fn(ctx)

	// Stop renewal and wait for it to fully exit before releasing, so release
	// never races an in-flight Renew against the same lock.
	cancel()
	renewDone.Wait()
	if renewErr := <-renewErrCh; renewErr != nil {
		opErr = errors.Join(opErr, fmt.Errorf("renew lock %q: %w", key, renewErr))
	}
	return opErr
}

func (s *Service) renew(ctx context.Context, lock *domain.Lock, ttl time.Duration) error {
	ticker := time.NewTicker(ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.manager.Renew(ctx, lock, ttl); err != nil {
				return err
			}
		}
	}
}
