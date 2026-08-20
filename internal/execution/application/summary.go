package application

import (
	"context"
	"fmt"

	"github.com/infra-orchestration/controlplane/internal/execution/domain"
)

type PlanSummary struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
}

func (s *Service) SummarizePlan(ctx context.Context, planID string) (PlanSummary, error) {
	tasks, err := s.repo.ListByPlan(ctx, planID)
	if err != nil {
		return PlanSummary{}, fmt.Errorf("list plan tasks: %w", err)
	}
	var summary PlanSummary
	summary.Total = len(tasks)
	for _, task := range tasks {
		switch task.Status {
		case domain.StatusPending, domain.StatusClaimed:
			summary.Pending++
		case domain.StatusRunning:
			summary.Running++
		case domain.StatusSucceeded:
			summary.Succeeded++
		case domain.StatusFailed:
			summary.Failed++
		case domain.StatusCancelled:
			summary.Cancelled++
		}
	}
	return summary, nil
}

func (s PlanSummary) Complete() bool {
	return s.Total > 0 && s.Pending == 0 && s.Running == 0
}

func (s PlanSummary) Successful() bool {
	return s.Complete() && s.Failed == 0 && s.Cancelled == 0
}
