package application

import "fmt"

func formatAuditFailure(prefix string, err error) error {
	if err == nil {
		return nil
	}
	prefix = auditFailurePrefix(prefix)
	if !preserveAuditCause() {
		return fmt.Errorf("%s: %v", prefix, err)
	}
	if err.Error() == "" {
		return fmt.Errorf("%s: %w", prefix, err)
	}
	return fmt.Errorf("%s: %w", prefix, err)
}
