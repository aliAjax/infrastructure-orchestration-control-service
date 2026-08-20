package domain

import (
	"context"
	"encoding/json"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusClaimed   Status = "claimed"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Task struct {
	ID            string          `json:"id"`
	PlanID        string          `json:"plan_id"`
	ResourceID    string          `json:"resource_id"`
	EnvironmentID string          `json:"environment_id"`
	Status        Status          `json:"status"`
	Attempt       int             `json:"attempt"`
	MaxAttempts   int             `json:"max_attempts"`
	Input         json.RawMessage `json:"input"`
	Output        json.RawMessage `json:"output"`
	Error         string          `json:"error"`
	LockKey       string          `json:"lock_key"`
	Timeout       int             `json:"timeout"`
	LeaseOwner    string          `json:"lease_owner"`
	CreatedAt     time.Time       `json:"created_at"`
	StartedAt     *time.Time      `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at"`
	NextRunAt     time.Time       `json:"next_run_at"`
}

type CreateTaskInput struct {
	ID            string
	PlanID        string
	ResourceID    string
	EnvironmentID string
	Input         json.RawMessage
	LockKey       string
	Timeout       int
	MaxAttempts   int
}

type RunnerRequest struct {
	ResourceID string          `json:"resource_id"`
	Type       string          `json:"type"`
	Input      json.RawMessage `json:"input"`
}

type RunnerResponse struct {
	Output json.RawMessage `json:"output"`
	Error  string          `json:"error"`
}

type Runner interface {
	Execute(ctx context.Context, request RunnerRequest) (RunnerResponse, error)
}

type ClaimedTask struct {
	Task  Task
	Token string
}
