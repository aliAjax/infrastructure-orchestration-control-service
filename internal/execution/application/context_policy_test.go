package application

import (
	"context"
	"errors"
	"testing"
)

func TestExecutionContextRejectsCancelledInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := executionContext(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
