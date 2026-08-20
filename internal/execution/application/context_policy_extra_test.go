package application

import (
	"context"
	"testing"
)

func TestRunnerContextIsCallerContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "k", "v")
	if runnerExecutionContext(ctx).Value("k") != "v" {
		t.Fatal("runner context detached")
	}
}
func TestStateContextIsCallerContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "k", "v")
	if stateUpdateContext(ctx).Value("k") != "v" {
		t.Fatal("state context detached")
	}
}
