package adapter

import (
	"encoding/json"

	"github.com/infra-orchestration/controlplane/internal/state/domain"
)

type StateDTO struct {
	ResourceID          string          `json:"resource_id"`
	EnvironmentID       string          `json:"environment_id"`
	DesiredState        json.RawMessage `json:"desired_state"`
	ActualState         json.RawMessage `json:"actual_state"`
	LastExecutionStatus string          `json:"last_execution_status"`
	Version             int             `json:"version"`
}

func StateFromDomain(s domain.ResourceState) StateDTO {
	return StateDTO{
		ResourceID:          s.ResourceID,
		EnvironmentID:       s.EnvironmentID,
		DesiredState:        s.DesiredState,
		ActualState:         s.ActualState,
		LastExecutionStatus: string(s.LastExecutionStatus),
		Version:             s.Version,
	}
}
