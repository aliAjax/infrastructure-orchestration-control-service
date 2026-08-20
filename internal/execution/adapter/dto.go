package adapter

import (
	"encoding/json"

	"github.com/infra-orchestration/controlplane/internal/execution/domain"
)

type TaskDTO struct {
	ID         string          `json:"id"`
	PlanID     string          `json:"plan_id"`
	ResourceID string          `json:"resource_id"`
	Status     string          `json:"status"`
	Attempt    int             `json:"attempt"`
	Input      json.RawMessage `json:"input"`
	Output     json.RawMessage `json:"output"`
	Error      string          `json:"error"`
}

func TaskFromDomain(t domain.Task) TaskDTO {
	return TaskDTO{ID: t.ID, PlanID: t.PlanID, ResourceID: t.ResourceID, Status: string(t.Status), Attempt: t.Attempt, Input: t.Input, Output: t.Output, Error: t.Error}
}
