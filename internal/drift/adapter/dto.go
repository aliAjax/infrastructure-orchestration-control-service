package adapter

import (
	"encoding/json"

	"github.com/infra-orchestration/controlplane/internal/drift/domain"
)

type DriftDTO struct {
	ID            string          `json:"id"`
	EnvironmentID string          `json:"environment_id"`
	ResourceID    string          `json:"resource_id"`
	DesiredState  json.RawMessage `json:"desired_state"`
	ActualState   json.RawMessage `json:"actual_state"`
	Severity      string          `json:"severity"`
	Resolved      bool            `json:"resolved"`
}

func DriftFromDomain(d domain.DriftRecord) DriftDTO {
	return DriftDTO{ID: d.ID, EnvironmentID: d.EnvironmentID, ResourceID: d.ResourceID, DesiredState: d.DesiredState, ActualState: d.ActualState, Severity: d.Severity, Resolved: d.Resolved}
}
