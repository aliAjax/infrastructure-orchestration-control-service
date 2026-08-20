package application

import plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"

func validateStatusUpdate(from, to plandomain.Status) error {
	return ValidatePlanTransition(from, to)
}
