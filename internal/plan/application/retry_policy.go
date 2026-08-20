package application

import plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"

func retryTransitionAllowed(from, to plandomain.Status) bool {
	return from == plandomain.StatusFailed && to == plandomain.StatusExecuting
}
