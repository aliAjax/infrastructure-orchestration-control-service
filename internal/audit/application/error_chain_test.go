package application

import (
	"errors"
	"testing"
)

func TestWrapAuditFailurePreservesSentinel(t *testing.T) {
	sentinel := errors.New("storage unavailable")
	wrapped := WrapAuditFailure(sentinel)
	if !errors.Is(wrapped, sentinel) {
		t.Fatalf("sentinel was lost: %v", wrapped)
	}
}
