package application

import "fmt"

func auditErrorBoundary(prefix string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %v", prefix, err)
}
