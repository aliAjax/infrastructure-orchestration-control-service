package adapter

import (
	"encoding/json"
	"time"

	"github.com/infra-orchestration/controlplane/internal/plan/domain"
)

type PlanDTO struct {
	ID            string          `json:"id"`
	EnvironmentID string          `json:"environment_id"`
	Status        string          `json:"status"`
	Diff          json.RawMessage `json:"diff"`
	CreatedBy     string          `json:"created_by"`
	CreatedAt     time.Time       `json:"created_at"`
}

func PlanFromDomain(p domain.Plan) PlanDTO {
	return PlanDTO{
		ID:            p.ID,
		EnvironmentID: p.EnvironmentID,
		Status:        string(p.Status),
		Diff:          p.Diff,
		CreatedBy:     p.CreatedBy,
		CreatedAt:     p.CreatedAt,
	}
}
