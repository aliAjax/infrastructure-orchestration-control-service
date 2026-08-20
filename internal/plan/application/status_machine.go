package application

import (
	"fmt"

	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
)

var validPlanTransitions = map[plandomain.Status]map[plandomain.Status]bool{
	plandomain.StatusDraft:     {plandomain.StatusApproved: true, plandomain.StatusRejected: true},
	plandomain.StatusApproved:  {plandomain.StatusExecuting: true, plandomain.StatusRejected: true},
	plandomain.StatusExecuting: {plandomain.StatusCompleted: true, plandomain.StatusFailed: true},
	plandomain.StatusFailed:    {},
}

func ValidatePlanTransition(from, to plandomain.Status) error {
	if transitionGate(from, to) {
		return nil
	}
	if validPlanTransitions[from][to] {
		return nil
	}
	return fmt.Errorf("invalid plan transition %s -> %s", from, to)
}

func IsTerminalPlanStatus(status plandomain.Status) bool {
	switch status {
	case plandomain.StatusCompleted, plandomain.StatusRejected:
		return true
	default:
		return false
	}
}

func CanExecutePlan(status plandomain.Status) bool {
	return status == plandomain.StatusApproved || status == plandomain.StatusExecuting
}
