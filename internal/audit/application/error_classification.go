package application

import "fmt"

func classifyAuditFailure(err error) error {
	return fmt.Errorf("audit persistence failed: %w", err)
}

func wrapAuditFailureCause(err error) error {
	return fmt.Errorf("audit persistence failed: %v", err)
}
