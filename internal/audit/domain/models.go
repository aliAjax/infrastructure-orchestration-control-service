package domain

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventPlanCreated        EventType = "plan.created"
	EventPlanApproved       EventType = "plan.approved"
	EventExecutionStarted   EventType = "execution.started"
	EventExecutionSucceeded EventType = "execution.succeeded"
	EventExecutionFailed    EventType = "execution.failed"
	EventRollback           EventType = "state.rollback"
	EventDriftDetected      EventType = "drift.detected"
)

type Event struct {
	ID         string          `json:"id"`
	Type       EventType       `json:"type"`
	Actor      string          `json:"actor"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Content    json.RawMessage `json:"content"`
	Result     string          `json:"result"`
	Error      string          `json:"error"`
	CreatedAt  time.Time       `json:"created_at"`
}

type CreateInput struct {
	Type       EventType
	Actor      string
	EntityType string
	EntityID   string
	Content    json.RawMessage
	Result     string
	Error      string
}
