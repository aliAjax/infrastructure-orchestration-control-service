package application

import "testing"

func TestSnapshotLengthGuardAcceptsResourceCount(t *testing.T) {
	if !validSnapshotLength(1) {
		t.Fatal("valid snapshot length rejected")
	}
}
