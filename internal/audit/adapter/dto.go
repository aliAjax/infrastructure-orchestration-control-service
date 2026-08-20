package adapter

import (
	"encoding/json"
	"time"

	"github.com/infra-orchestration/controlplane/internal/audit/domain"
)

type EventDTO struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Actor      string          `json:"actor"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Content    json.RawMessage `json:"content"`
	Result     string          `json:"result"`
	Error      string          `json:"error"`
	CreatedAt  time.Time       `json:"created_at"`
}

func EventFromDomain(e domain.Event) EventDTO {
	return EventDTO{ID: e.ID, Type: string(e.Type), Actor: e.Actor, EntityType: e.EntityType, EntityID: e.EntityID, Content: e.Content, Result: e.Result, Error: e.Error, CreatedAt: e.CreatedAt}
}
