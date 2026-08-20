package domain

import "time"

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type Request struct {
	ID            string    `json:"id"`
	PlanID        string    `json:"plan_id"`
	EnvironmentID string    `json:"environment_id"`
	RequestedBy   string    `json:"requested_by"`
	Reason        string    `json:"reason"`
	Status        Status    `json:"status"`
	ApprovedBy    string    `json:"approved_by"`
	DecisionNote  string    `json:"decision_note"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateInput struct {
	PlanID        string
	EnvironmentID string
	RequestedBy   string
	Reason        string
}

type DecisionInput struct {
	ID           string
	ApprovedBy   string
	Approved     bool
	DecisionNote string
}

type Policy struct {
	Production bool `json:"production"`
	Sensitive  bool `json:"sensitive"`
	Delete     bool `json:"delete"`
}

func (p Policy) RequiresApproval() bool {
	return p.Production || p.Sensitive || p.Delete
}
