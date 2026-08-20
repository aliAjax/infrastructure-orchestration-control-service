package domain

import (
	"context"
	"time"
)

type Lock struct {
	Key     string
	Token   string
	Expires time.Time
}

type Manager interface {
	Acquire(ctx context.Context, key, owner string, ttl time.Duration) (*Lock, error)
	Release(ctx context.Context, lock *Lock) error
	Renew(ctx context.Context, lock *Lock, ttl time.Duration) error
}

type NoopManager struct{}

func (NoopManager) Acquire(ctx context.Context, key, owner string, ttl time.Duration) (*Lock, error) {
	return &Lock{Key: key, Token: owner, Expires: time.Now().Add(ttl)}, nil
}

func (NoopManager) Release(ctx context.Context, lock *Lock) error {
	return nil
}

func (NoopManager) Renew(ctx context.Context, lock *Lock, ttl time.Duration) error {
	lock.Expires = time.Now().Add(ttl)
	return nil
}
