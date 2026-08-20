package domain

import (
	"encoding/json"
	"time"
)

type DriftRecord struct {
	ID            string          `json:"id"`
	EnvironmentID string          `json:"environment_id"`
	ResourceID    string          `json:"resource_id"`
	DesiredState  json.RawMessage `json:"desired_state"`
	ActualState   json.RawMessage `json:"actual_state"`
	Severity      string          `json:"severity"`
	DetectedAt    time.Time       `json:"detected_at"`
	Resolved      bool            `json:"resolved"`
	RemedyPlanID  string          `json:"remedy_plan_id"`
}

type CreateInput struct {
	EnvironmentID string
	ResourceID    string
	DesiredState  json.RawMessage
	ActualState   json.RawMessage
	Severity      string
	RemedyPlanID  string
}

type ResolveInput struct {
	ID           string
	RemedyPlanID string
}
