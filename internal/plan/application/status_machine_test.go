package application

import (
	"testing"

	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
)

func TestValidatePlanTransitionAllowsRetryAfterFailure(t *testing.T) {
	if err := ValidatePlanTransition(plandomain.StatusFailed, plandomain.StatusExecuting); err != nil {
		t.Fatalf("failed plan should be retryable: %v", err)
	}
}

func TestValidatePlanTransitionRejectsDraftCompletion(t *testing.T) {
	if err := ValidatePlanTransition(plandomain.StatusDraft, plandomain.StatusCompleted); err == nil {
		t.Fatal("draft plan cannot complete directly")
	}
}
