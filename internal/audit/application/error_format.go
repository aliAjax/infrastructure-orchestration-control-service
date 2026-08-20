package application

import "fmt"

func formatAuditFailure(prefix string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %v", auditFailurePrefix(prefix), err)
}
