package application

import (
	"context"
	"encoding/json"
	"fmt"

	graphapp "github.com/infra-orchestration/controlplane/internal/graph/application"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
	resourcedomain "github.com/infra-orchestration/controlplane/internal/resource/domain"
	statedomain "github.com/infra-orchestration/controlplane/internal/state/domain"
)

type Service struct {
	repo         plandomain.Repository
	resourceRepo resourceRepository
	stateRepo    stateRepository
	graphService *graphapp.Service
}

type resourceRepository interface {
	ListResources(ctx context.Context, environmentID string) ([]resourcedomain.Resource, error)
}

type stateRepository interface {
	ListByEnvironment(ctx context.Context, environmentID string) ([]statedomain.ResourceState, error)
}

func NewService(repo plandomain.Repository, resources resourceRepository, states stateRepository, graphService *graphapp.Service) *Service {
	return &Service{repo: repo, resourceRepo: resources, stateRepo: states, graphService: graphService}
}

func (s *Service) Generate(ctx context.Context, input plandomain.GenerateInput) (plandomain.Plan, error) {
	if input.EnvironmentID == "" {
		return plandomain.Plan{}, fmt.Errorf("environment_id is required")
	}
	resources, err := s.resourceRepo.ListResources(ctx, input.EnvironmentID)
	if err != nil {
		return plandomain.Plan{}, fmt.Errorf("load resources: %w", err)
	}
	if err := s.graphService.ValidateDependencies(ctx, resources); err != nil {
		return plandomain.Plan{}, fmt.Errorf("dependency graph invalid: %w", err)
	}
	states, err := s.stateRepo.ListByEnvironment(ctx, input.EnvironmentID)
	if err != nil {
		return plandomain.Plan{}, fmt.Errorf("load states: %w", err)
	}
	stateMap := make(map[string]statedomain.ResourceState, len(states))
	for _, st := range states {
		stateMap[st.ResourceID] = st
	}
	items := make([]plandomain.DiffItem, 0, len(resources))
	for _, res := range resources {
		current, ok := stateMap[res.ID]
		before := []byte(`{}`)
		if ok && len(current.ActualState) > 0 {
			before = current.ActualState
		}
		op := plandomain.OperationCreate
		if ok && len(current.ActualState) > 0 {
			op = plandomain.OperationUpdate
		}
		items = append(items, plandomain.DiffItem{
			ResourceID: res.ID,
			Name:       res.Name,
			Type:       res.Type,
			Operation:  op,
			Before:     before,
			After:      res.DesiredState,
		})
	}
	diff, err := json.Marshal(items)
	if err != nil {
		return plandomain.Plan{}, fmt.Errorf("marshal diff: %w", err)
	}
	plan, err := s.repo.Create(ctx, plandomain.CreateInput{EnvironmentID: input.EnvironmentID, Status: plandomain.StatusDraft, Diff: diff, CreatedBy: input.CreatedBy})
	if err != nil {
		return plandomain.Plan{}, err
	}
	if err := s.repo.SaveDiffItems(ctx, plan.ID, items); err != nil {
		return plandomain.Plan{}, fmt.Errorf("save diff items: %w", err)
	}
	return plan, nil
}

func (s *Service) Get(ctx context.Context, id string) (plandomain.Plan, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, environmentID string, limit int) ([]plandomain.Plan, error) {
	return s.repo.ListByEnvironment(ctx, environmentID, limit)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status plandomain.Status) (plandomain.Plan, error) {
	if id == "" {
		return plandomain.Plan{}, fmt.Errorf("plan id is required")
	}
	if status == "" {
		return plandomain.Plan{}, fmt.Errorf("plan status is required")
	}
	plan, err := s.repo.Get(ctx, id)
	if err != nil {
		return plandomain.Plan{}, err
	}
	if err := ValidatePlanTransition(plan.Status, status); err != nil {
		return plandomain.Plan{}, err
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) DiffItems(ctx context.Context, id string) ([]plandomain.DiffItem, error) {
	return s.repo.GetDiffItems(ctx, id)
}
