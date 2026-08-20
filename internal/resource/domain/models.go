package domain

import (
	"encoding/json"
	"time"
)

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Environment struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	Name       string    `json:"name"`
	Kind       string    `json:"kind"`
	Production bool      `json:"production"`
	LockKey    string    `json:"lock_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Resource struct {
	ID               string          `json:"id"`
	EnvironmentID    string          `json:"environment_id"`
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	Provider         string          `json:"provider"`
	DesiredState     json.RawMessage `json:"desired_state"`
	DependsOn        []string        `json:"depends_on"`
	Version          int             `json:"version"`
	Sensitive        bool            `json:"sensitive"`
	ApprovalRequired bool            `json:"approval_required"`
	LockKey          string          `json:"lock_key"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type Variable struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Sensitive bool      `json:"sensitive"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EnvironmentInput struct {
	ProjectID  string `json:"project_id"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Production bool   `json:"production"`
	LockKey    string `json:"lock_key"`
}

type ResourceInput struct {
	EnvironmentID    string          `json:"environment_id"`
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	Provider         string          `json:"provider"`
	DesiredState     json.RawMessage `json:"desired_state"`
	DependsOn        []string        `json:"depends_on"`
	Sensitive        bool            `json:"sensitive"`
	ApprovalRequired bool            `json:"approval_required"`
	LockKey          string          `json:"lock_key"`
}

type VariableInput struct {
	ProjectID string `json:"project_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive"`
}
