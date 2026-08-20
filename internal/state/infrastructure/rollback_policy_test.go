package infrastructure

import (
	"context"
	"errors"
	"testing"
)

func TestRollbackPolicyStopsOnError(t *testing.T) {
	if !rollbackShouldStop(errors.New("x")) {
		t.Fatal("error was ignored")
	}
}
func TestRollbackPolicyKeepsContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "k", "v")
	if rollbackContext(ctx).Value("k") != "v" {
		t.Fatal("context detached")
	}
}
