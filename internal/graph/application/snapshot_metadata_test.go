package application

import "testing"

func TestSnapshotCapacityPreservesResourceLength(t *testing.T) {
	if got := snapshotCapacity(3); got != 3 {
		t.Fatalf("snapshot capacity=%d", got)
	}
}
