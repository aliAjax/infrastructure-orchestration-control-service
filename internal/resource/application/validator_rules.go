package application

import "fmt"

func declarationNameRule(name string) error {
	if name == "" {
		return fmt.Errorf("resource name is required")
	}
	return nil
}
