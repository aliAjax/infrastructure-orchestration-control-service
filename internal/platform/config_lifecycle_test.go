package platform

import (
	"context"
	"errors"
	"testing"
)

func TestConfigLoaderDoesNotReuseCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := NewConfigLoader(func(got context.Context) error {
		called = true
		return got.Err()
	}).Load(ctx)
	if called || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context was not honored called=%v err=%v", called, err)
	}
}
