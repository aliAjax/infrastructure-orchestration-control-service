package domain

import (
	"encoding/json"
	"time"
)

type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusSucceeded ExecutionStatus = "succeeded"
	StatusFailed    ExecutionStatus = "failed"
)

type ResourceState struct {
	ID                  string          `json:"id"`
	ResourceID          string          `json:"resource_id"`
	EnvironmentID       string          `json:"environment_id"`
	DesiredState        json.RawMessage `json:"desired_state"`
	ActualState         json.RawMessage `json:"actual_state"`
	LastExecutionStatus ExecutionStatus `json:"last_execution_status"`
	Version             int             `json:"version"`
	LockVersion         int             `json:"lock_version"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type StateSnapshot struct {
	ID           string          `json:"id"`
	ResourceID   string          `json:"resource_id"`
	DesiredState json.RawMessage `json:"desired_state"`
	ActualState  json.RawMessage `json:"actual_state"`
	Status       ExecutionStatus `json:"status"`
	Version      int             `json:"version"`
	CapturedAt   time.Time       `json:"captured_at"`
}

type DesiredUpdate struct {
	ResourceID   string          `json:"resource_id"`
	DesiredState json.RawMessage `json:"desired_state"`
}

type ActualUpdate struct {
	ResourceID  string          `json:"resource_id"`
	ActualState json.RawMessage `json:"actual_state"`
	Status      ExecutionStatus `json:"status"`
}
