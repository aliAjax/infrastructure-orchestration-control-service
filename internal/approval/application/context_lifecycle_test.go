package application

import (
	"context"
	"errors"
	"testing"

	"github.com/infra-orchestration/controlplane/internal/approval/domain"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
)

type contextKey string

type contextProbeRepository struct {
	getCalls    int
	listCalls   int
	decideCalls int
	writeValue  any
}

func (r *contextProbeRepository) Create(context.Context, domain.CreateInput) (domain.Request, error) {
	return domain.Request{}, nil
}

func (r *contextProbeRepository) Get(context.Context, string) (domain.Request, error) {
	r.getCalls++
	return domain.Request{ID: "approval-1", Status: domain.StatusPending}, nil
}

func (r *contextProbeRepository) ListByPlan(context.Context, string) ([]domain.Request, error) {
	r.listCalls++
	return []domain.Request{{ID: "approval-1", Status: domain.StatusApproved}}, nil
}

func (r *contextProbeRepository) Decide(ctx context.Context, _ domain.DecisionInput) (domain.Request, error) {
	r.decideCalls++
	r.writeValue = ctx.Value(contextKey("request"))
	return domain.Request{ID: "approval-1", Status: domain.StatusApproved}, nil
}

func TestDecideRejectsCancelledContextBeforeRead(t *testing.T) {
	repo := &contextProbeRepository{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewService(repo).Decide(ctx, domain.DecisionInput{ID: "approval-1", ApprovedBy: "operator"})
	if !errors.Is(err, context.Canceled) || repo.getCalls != 0 || repo.decideCalls != 0 {
		t.Fatalf("cancelled decision reached storage: get=%d decide=%d err=%v", repo.getCalls, repo.decideCalls, err)
	}
}

func TestDecidePassesCallerContextToWrite(t *testing.T) {
	repo := &contextProbeRepository{}
	ctx := context.WithValue(context.Background(), contextKey("request"), "trace-17")
	_, err := NewService(repo).Decide(ctx, domain.DecisionInput{ID: "approval-1", ApprovedBy: "operator"})
	if err != nil || repo.writeValue != "trace-17" {
		t.Fatalf("decision lost caller context: value=%v err=%v", repo.writeValue, err)
	}
}

func TestEnsureApprovedRejectsCancelledContextBeforeList(t *testing.T) {
	repo := &contextProbeRepository{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := NewService(repo).EnsureApproved(ctx, "plan-1")
	if !errors.Is(err, context.Canceled) || repo.listCalls != 0 {
		t.Fatalf("cancelled approval check reached storage: list=%d err=%v", repo.listCalls, err)
	}
}

func TestPolicyEvaluatorRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewPolicyEvaluator(nil, nil).Evaluate(ctx, "env-1", []plandomain.DiffItem{{ResourceID: "r1"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("policy evaluation ignored cancellation: %v", err)
	}
}

func TestValidateApprovalStateRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ValidateApprovalState(ctx, domain.Policy{Production: true}, []domain.Request{{ID: "approval-1", Status: domain.StatusApproved}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("approval validation ignored cancellation: %v", err)
	}
}
