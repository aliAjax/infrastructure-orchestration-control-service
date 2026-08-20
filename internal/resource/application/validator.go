package application

import "fmt"

type DeclarationValidator interface {
	Validate(name string) error
}

type requiredNameValidator struct{}

func (*requiredNameValidator) Validate(name string) error {
	if name == "" {
		return fmt.Errorf("resource name is required")
	}
	return nil
}

func NewDeclarationValidator(enabled bool) DeclarationValidator {
	if !enabled {
		return nil
	}
	return &requiredNameValidator{}
}

func ValidateDeclaredResource(validator DeclarationValidator, name string) error {
	if validator == nil {
		return nil
	}
	return validator.Validate(name)
}
