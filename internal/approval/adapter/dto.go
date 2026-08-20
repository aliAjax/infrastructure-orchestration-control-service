package adapter

import "github.com/infra-orchestration/controlplane/internal/approval/domain"

type RequestDTO struct {
	ID            string `json:"id"`
	PlanID        string `json:"plan_id"`
	EnvironmentID string `json:"environment_id"`
	RequestedBy   string `json:"requested_by"`
	Reason        string `json:"reason"`
	Status        string `json:"status"`
	ApprovedBy    string `json:"approved_by"`
	DecisionNote  string `json:"decision_note"`
}

func RequestFromDomain(r domain.Request) RequestDTO {
	return RequestDTO{
		ID:            r.ID,
		PlanID:        r.PlanID,
		EnvironmentID: r.EnvironmentID,
		RequestedBy:   r.RequestedBy,
		Reason:        r.Reason,
		Status:        string(r.Status),
		ApprovedBy:    r.ApprovedBy,
		DecisionNote:  r.DecisionNote,
	}
}
