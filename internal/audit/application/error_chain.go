package application

import "fmt"

// WrapAuditFailure preserves the repository error for retry and classification.
func WrapAuditFailure(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("audit persistence failed: %w", err)
}
