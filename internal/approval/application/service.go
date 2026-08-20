package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/infra-orchestration/controlplane/internal/approval/domain"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input domain.CreateInput) (domain.Request, error) {
	input.Reason = strings.TrimSpace(input.Reason)
	if input.PlanID == "" || input.EnvironmentID == "" || input.RequestedBy == "" {
		return domain.Request{}, fmt.Errorf("plan_id, environment_id and requested_by are required")
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Decide(ctx context.Context, input domain.DecisionInput) (domain.Request, error) {
	if input.ID == "" || input.ApprovedBy == "" {
		return domain.Request{}, fmt.Errorf("id and approved_by are required")
	}
	request, err := s.repo.Get(ctx, input.ID)
	if err != nil {
		return domain.Request{}, err
	}
	if request.Status != domain.StatusPending {
		return domain.Request{}, fmt.Errorf("approval request is not pending")
	}
	return s.repo.Decide(ctx, input)
}

func (s *Service) ForPlan(ctx context.Context, planID string) ([]domain.Request, error) {
	return s.repo.ListByPlan(ctx, planID)
}

func (s *Service) EnsureApproved(ctx context.Context, planID string) error {
	requests, err := s.repo.ListByPlan(ctx, planID)
	if err != nil {
		return fmt.Errorf("list approvals: %w", err)
	}
	if len(requests) == 0 {
		return nil
	}
	for _, request := range requests {
		if request.Status != domain.StatusApproved {
			return fmt.Errorf("plan %s has unapproved approval request %s", planID, request.ID)
		}
	}
	return nil
}

func (s *Service) PolicyForDiff(items []plandomain.DiffItem) domain.Policy {
	policy := domain.Policy{}
	for _, item := range items {
		if item.Operation == plandomain.OperationDelete {
			policy.Delete = true
		}
	}
	return policy
}
