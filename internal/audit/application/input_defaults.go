package application

import "github.com/infra-orchestration/controlplane/internal/audit/domain"

func applyAuditInputDefaults(input domain.CreateInput) domain.CreateInput {
	if input.EntityType == "" {
		input.EntityType = "unknown"
	}
	if input.EntityID == "" {
		input.EntityID = "unassigned"
	}
	if input.Content == nil {
		input.Content = []byte(`{}`)
	}
	return input
}
