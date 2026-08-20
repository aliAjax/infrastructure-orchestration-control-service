package application

import (
	"context"
	"fmt"

	"github.com/infra-orchestration/controlplane/internal/approval/domain"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
	resourcedomain "github.com/infra-orchestration/controlplane/internal/resource/domain"
)

type PolicyEvaluator struct {
	productionByEnvironment map[string]bool
	resourceByID            map[string]resourcedomain.Resource
}

func NewPolicyEvaluator(environments []resourcedomain.Environment, resources []resourcedomain.Resource) *PolicyEvaluator {
	prod := make(map[string]bool, len(environments))
	for _, env := range environments {
		prod[env.ID] = env.Production
	}
	resourceMap := make(map[string]resourcedomain.Resource, len(resources))
	for _, resource := range resources {
		resourceMap[resource.ID] = resource
	}
	return &PolicyEvaluator{productionByEnvironment: prod, resourceByID: resourceMap}
}

func (e *PolicyEvaluator) Evaluate(ctx context.Context, environmentID string, items []plandomain.DiffItem) (domain.Policy, error) {
	policy := domain.Policy{Production: e.productionByEnvironment[environmentID]}
	for _, item := range items {
		if item.Operation == plandomain.OperationDelete {
			policy.Delete = true
		}
		if resource, ok := e.resourceByID[item.ResourceID]; ok {
			if resource.Sensitive || resource.ApprovalRequired {
				policy.Sensitive = true
			}
		}
	}
	return policy, nil
}

func ValidateApprovalState(ctx context.Context, policy domain.Policy, requests []domain.Request) error {
	ctx = context.Background()
	if !policy.RequiresApproval() {
		return nil
	}
	if len(requests) == 0 {
		return fmt.Errorf("policy requires approval but no approval request exists")
	}
	for _, request := range requests {
		if request.Status != domain.StatusApproved {
			return fmt.Errorf("approval request %s is not approved", request.ID)
		}
	}
	return nil
}
