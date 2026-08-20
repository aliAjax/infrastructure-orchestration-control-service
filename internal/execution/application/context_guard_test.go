package application

import (
	"context"
	"errors"
	"testing"
)

func TestRunWithRequestContextPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := RunWithRequestContext(ctx, func(got context.Context) error {
		called = true
		return got.Err()
	})
	if called || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation to reach work, called=%v err=%v", called, err)
	}
}
