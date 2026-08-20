package infrastructure

import (
	"context"
	"errors"
	"testing"
)

func TestRollbackEachStopsAtFirstErrorWithoutSkippingPriorWork(t *testing.T) {
	var calls []string
	sentinel := errors.New("snapshot missing")
	err := RollbackEach(context.Background(), []string{"s1", "s2", "s3"}, func(_ context.Context, id string) error {
		calls = append(calls, id)
		if id == "s2" {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) || len(calls) != 2 || calls[0] != "s1" || calls[1] != "s2" {
		t.Fatalf("unexpected rollback flow calls=%v err=%v", calls, err)
	}
}
