package application

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/infra-orchestration/controlplane/internal/lock/domain"
)

type cleanupKey string

type lifecycleManager struct {
	acquireCalls int
	releaseErr   error
	renewErr     error
	renewDelay   time.Duration
	renewStarted chan struct{}
	renewDone    chan struct{}
	startOnce    sync.Once
	doneOnce     sync.Once
	renewExited  atomic.Bool
	earlyRelease atomic.Bool
	releaseValue any
	nilLock      bool
}

type concurrentLifecycleManager struct {
	activeRenewals atomic.Int32
	activeByToken  sync.Map
	startedByToken sync.Map
	earlyRelease   atomic.Bool
	renewStarted   chan struct{}
}

func (m *concurrentLifecycleManager) Acquire(_ context.Context, _ string, owner string, _ time.Duration) (*domain.Lock, error) {
	return &domain.Lock{Key: "deployment", Token: owner, Expires: time.Now().Add(time.Second)}, nil
}

func (m *concurrentLifecycleManager) Renew(ctx context.Context, lock *domain.Lock, _ time.Duration) error {
	m.activeRenewals.Add(1)
	m.activeByToken.Store(lock.Token, true)
	if _, loaded := m.startedByToken.LoadOrStore(lock.Token, true); !loaded {
		m.renewStarted <- struct{}{}
	}
	<-ctx.Done()
	time.Sleep(40 * time.Millisecond)
	m.activeByToken.Delete(lock.Token)
	m.activeRenewals.Add(-1)
	return nil
}

func (m *concurrentLifecycleManager) Release(_ context.Context, lock *domain.Lock) error {
	if _, active := m.activeByToken.Load(lock.Token); active {
		m.earlyRelease.Store(true)
	}
	return nil
}

func (m *lifecycleManager) Acquire(context.Context, string, string, time.Duration) (*domain.Lock, error) {
	m.acquireCalls++
	if m.nilLock {
		return nil, nil
	}
	return &domain.Lock{Key: "deployment", Token: "owner-1", Expires: time.Now().Add(time.Second)}, nil
}

func (m *lifecycleManager) Renew(ctx context.Context, _ *domain.Lock, _ time.Duration) error {
	if m.renewStarted != nil {
		m.startOnce.Do(func() { close(m.renewStarted) })
	}
	if m.renewErr != nil {
		m.renewExited.Store(true)
		m.doneOnce.Do(func() {
			if m.renewDone != nil {
				close(m.renewDone)
			}
		})
		return m.renewErr
	}
	<-ctx.Done()
	if m.renewDelay > 0 {
		time.Sleep(m.renewDelay)
	}
	m.renewExited.Store(true)
	m.doneOnce.Do(func() {
		if m.renewDone != nil {
			close(m.renewDone)
		}
	})
	return nil
}

func (m *lifecycleManager) Release(ctx context.Context, _ *domain.Lock) error {
	if m.renewStarted != nil && !m.renewExited.Load() {
		m.earlyRelease.Store(true)
	}
	m.releaseValue = ctx.Value(cleanupKey("trace"))
	return m.releaseErr
}

func finishWithLock(t *testing.T, call func() error) error {
	t.Helper()
	result := make(chan error, 1)
	go func() { result <- call() }()
	select {
	case err := <-result:
		return err
	case <-time.After(10 * time.Second):
		return errors.New("lock operation did not finish")
	}
}

func TestWithLockWaitsForRenewalExitBeforeRelease(t *testing.T) {
	m := &concurrentLifecycleManager{renewStarted: make(chan struct{}, 2)}
	service := NewService(m)
	start := make(chan struct{})
	leaveCriticalSection := make(chan struct{})
	errs := make(chan error, 2)
	var workers sync.WaitGroup
	for i := 0; i < 2; i++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			<-start
			errs <- service.WithLock(context.Background(), "deployment", string(rune('a'+worker)), 10*time.Millisecond, func(context.Context) error {
				<-leaveCriticalSection
				return nil
			})
		}(i)
	}
	close(start)
	for i := 0; i < 2; i++ {
		select {
		case <-m.renewStarted:
		case <-time.After(time.Second):
			t.Fatal("both renewals did not start")
		}
	}
	close(leaveCriticalSection)
	workersDone := make(chan struct{})
	go func() {
		workers.Wait()
		close(workersDone)
	}()
	select {
	case <-workersDone:
	case <-time.After(10 * time.Second):
		t.Fatal("lock operations did not finish after renewal cancellation")
	}
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent lock operation failed: %v", err)
		}
	}
	if m.earlyRelease.Load() || m.activeRenewals.Load() != 0 {
		t.Fatalf("release raced renewal exit: early=%v active=%d", m.earlyRelease.Load(), m.activeRenewals.Load())
	}
}

func TestWithLockReturnsReleaseError(t *testing.T) {
	releaseErr := errors.New("release unavailable")
	err := finishWithLock(t, func() error {
		return NewService(&lifecycleManager{releaseErr: releaseErr}).WithLock(context.Background(), "deployment", "worker", time.Second, func(context.Context) error { return nil })
	})
	if !errors.Is(err, releaseErr) {
		t.Fatalf("release error was lost: %v", err)
	}
}

func TestWithLockJoinsOperationAndReleaseErrors(t *testing.T) {
	operationErr := errors.New("operation failed")
	releaseErr := errors.New("release failed")
	err := finishWithLock(t, func() error {
		return NewService(&lifecycleManager{releaseErr: releaseErr}).WithLock(context.Background(), "deployment", "worker", time.Second, func(context.Context) error { return operationErr })
	})
	if !errors.Is(err, operationErr) || !errors.Is(err, releaseErr) {
		t.Fatalf("cleanup did not preserve both errors: %v", err)
	}
}

func TestWithLockCleanupKeepsValuesAfterCancellation(t *testing.T) {
	m := &lifecycleManager{}
	base := context.WithValue(context.Background(), cleanupKey("trace"), "request-42")
	ctx, cancel := context.WithCancel(base)
	err := finishWithLock(t, func() error {
		return NewService(m).WithLock(ctx, "deployment", "worker", time.Second, func(context.Context) error {
			cancel()
			return nil
		})
	})
	if err != nil || m.releaseValue != "request-42" {
		t.Fatalf("cleanup context lost request value: value=%v err=%v", m.releaseValue, err)
	}
}

func TestWithLockRejectsNonPositiveTTLBeforeAcquire(t *testing.T) {
	m := &lifecycleManager{}
	err := finishWithLock(t, func() error {
		return NewService(m).WithLock(context.Background(), "deployment", "worker", 0, func(context.Context) error { return nil })
	})
	if err == nil || m.acquireCalls != 0 {
		t.Fatalf("invalid ttl reached manager: calls=%d err=%v", m.acquireCalls, err)
	}
}

func TestWithLockRejectsNilAcquiredLock(t *testing.T) {
	m := &lifecycleManager{nilLock: true}
	called := false
	err := finishWithLock(t, func() error {
		return NewService(m).WithLock(context.Background(), "deployment", "worker", time.Second, func(context.Context) error {
			called = true
			return nil
		})
	})
	if err == nil || called {
		t.Fatalf("nil lock entered operation: called=%v err=%v", called, err)
	}
}

func TestWithLockReturnsRenewalError(t *testing.T) {
	renewErr := errors.New("renew unavailable")
	m := &lifecycleManager{renewErr: renewErr, renewStarted: make(chan struct{}), renewDone: make(chan struct{})}
	err := finishWithLock(t, func() error {
		return NewService(m).WithLock(context.Background(), "deployment", "worker", 10*time.Millisecond, func(context.Context) error {
			select {
			case <-m.renewDone:
				return nil
			case <-time.After(time.Second):
				return errors.New("renewal did not finish")
			}
		})
	})
	if !errors.Is(err, renewErr) {
		t.Fatalf("renewal error was lost: %v", err)
	}
}
