package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/infra-orchestration/controlplane/internal/execution/domain"
	"github.com/infra-orchestration/controlplane/internal/plan/application"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
	statedomain "github.com/infra-orchestration/controlplane/internal/state/domain"
)

type Service struct {
	repo         domain.Repository
	runner       domain.Runner
	planService  *application.Service
	stateService stateApplication
}

type stateApplication interface {
	ReportActual(ctx context.Context, resourceID string, actual json.RawMessage, status statedomain.ExecutionStatus) (statedomain.ResourceState, error)
}

func NewService(repo domain.Repository, runner domain.Runner, planService *application.Service, stateService stateApplication) *Service {
	return &Service{repo: repo, runner: runner, planService: planService, stateService: stateService}
}

func (s *Service) StartPlan(ctx context.Context, planID string, maxAttempts int) ([]domain.Task, error) {
	plan, err := s.planService.Get(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("load plan: %w", err)
	}
	if plan.Status != plandomain.StatusApproved {
		return nil, fmt.Errorf("plan %s is not approved", planID)
	}
	items, err := s.planService.DiffItems(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("load diff items: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("plan has no diff items")
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	inputs := make([]domain.CreateTaskInput, 0, len(items))
	for _, item := range items {
		input, _ := json.Marshal(map[string]any{"operation": item.Operation, "resource_id": item.ResourceID, "name": item.Name, "type": item.Type, "desired": item.After})
		inputs = append(inputs, domain.CreateTaskInput{
			ID:            uuid.NewString(),
			PlanID:        planID,
			ResourceID:    item.ResourceID,
			EnvironmentID: plan.EnvironmentID,
			Input:         input,
			LockKey:       "plan:" + plan.EnvironmentID,
			Timeout:       30000,
			MaxAttempts:   maxAttempts,
		})
	}
	tasks, err := s.repo.CreateTasks(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("create tasks: %w", err)
	}
	if _, err := s.planService.UpdateStatus(ctx, planID, plandomain.StatusExecuting); err != nil {
		return nil, fmt.Errorf("mark plan executing: %w", err)
	}
	return tasks, nil
}

func (s *Service) ProcessOne(ctx context.Context, owner string) (*domain.Task, error) {
	task, err := s.repo.ClaimNext(ctx, owner)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}
	timeout := executionTimeout(int64(task.Timeout))
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var request domain.RunnerRequest
	if err := json.Unmarshal(task.Input, &request); err != nil {
		_, _ = s.repo.Fail(ctx, task.ID, owner, "invalid runner request")
		return task, err
	}
	response, err := s.runner.Execute(runnerExecutionContext(runCtx), request)
	if err != nil {
		_, _ = s.repo.Fail(ctx, task.ID, owner, err.Error())
		return task, err
	}
	if response.Error != "" {
		_, _ = s.repo.Fail(ctx, task.ID, owner, response.Error)
		if s.stateService != nil {
			_, _ = s.stateService.ReportActual(stateUpdateContext(ctx), task.ResourceID, request.Input, statedomain.StatusFailed)
		}
		return task, fmt.Errorf("runner failed: %s", response.Error)
	}
	completed, err := s.repo.Complete(ctx, task.ID, owner, response.Output)
	if err != nil {
		return task, err
	}
	if s.stateService != nil {
		_, _ = s.stateService.ReportActual(stateUpdateContext(ctx), task.ResourceID, response.Output, statedomain.StatusSucceeded)
	}
	return &completed, nil
}

func (s *Service) List(ctx context.Context, planID string) ([]domain.Task, error) {
	return s.repo.ListByPlan(ctx, planID)
}

func (s *Service) Retry(ctx context.Context, taskID string) (domain.Task, error) {
	return s.repo.Retry(ctx, taskID)
}

func (s *Service) Cancel(ctx context.Context, taskID string) (domain.Task, error) {
	return s.repo.Cancel(ctx, taskID)
}
