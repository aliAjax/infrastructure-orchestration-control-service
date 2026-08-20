package application

import (
	"errors"
	"testing"
)

func TestScanErrorPolicyCollectsFailures(t *testing.T) {
	if !shouldCollectScanError(errors.New("timeout")) {
		t.Fatal("scan failure was dropped")
	}
}

func TestScanWorkerPolicyCountsAllEnvironments(t *testing.T) {
	if scanWorkerCount([]string{"a", "b"}) != 2 {
		t.Fatal("worker count mismatch")
	}
}
