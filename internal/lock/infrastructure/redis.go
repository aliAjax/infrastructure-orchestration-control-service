package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/infra-orchestration/controlplane/internal/lock/domain"
)

type RedisManager struct {
	client *redis.Client
}

func NewRedisManager(client *redis.Client) *RedisManager {
	return &RedisManager{client: client}
}

func (m *RedisManager) Acquire(ctx context.Context, key, owner string, ttl time.Duration) (*domain.Lock, error) {
	token := uuid.NewString()
	ok, err := m.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis setnx: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("lock %q is held by another owner", key)
	}
	return &domain.Lock{Key: key, Token: token, Expires: time.Now().Add(ttl)}, nil
}

func (m *RedisManager) Release(ctx context.Context, lock *domain.Lock) error {
	script := `if redis.call("get",KEYS[1])==ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end`
	if err := m.client.Eval(ctx, script, []string{lock.Key}, lock.Token).Err(); err != nil {
		return fmt.Errorf("redis release: %w", err)
	}
	return nil
}

func (m *RedisManager) Renew(ctx context.Context, lock *domain.Lock, ttl time.Duration) error {
	script := `if redis.call("get",KEYS[1])==ARGV[1] then return redis.call("expire",KEYS[1],ARGV[2]) else return 0 end`
	if err := m.client.Eval(ctx, script, []string{lock.Key}, lock.Token, int(ttl.Seconds())).Err(); err != nil {
		return fmt.Errorf("redis renew: %w", err)
	}
	lock.Expires = time.Now().Add(ttl)
	return nil
}
