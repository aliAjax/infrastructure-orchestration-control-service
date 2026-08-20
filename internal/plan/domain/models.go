package domain

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
	StatusExecuting Status = "executing"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Operation string

const (
	OperationCreate Operation = "create"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
	OperationDrift  Operation = "drift"
)

type Plan struct {
	ID            string          `json:"id"`
	EnvironmentID string          `json:"environment_id"`
	Status        Status          `json:"status"`
	Diff          json.RawMessage `json:"diff"`
	CreatedBy     string          `json:"created_by"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type DiffItem struct {
	ResourceID string          `json:"resource_id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Operation  Operation       `json:"operation"`
	Before     json.RawMessage `json:"before"`
	After      json.RawMessage `json:"after"`
}

type GenerateInput struct {
	EnvironmentID string
	CreatedBy     string
}

type CreateInput struct {
	ID            string
	EnvironmentID string
	Status        Status
	Diff          json.RawMessage
	CreatedBy     string
}
