package application

import "fmt"

func buildDeclarationValidator(enabled bool) DeclarationValidator {
	return declarationValidatorBuilder(enabled)
}

func validateRequiredName(name string) error {
	if !declarationValidationGate() {
		return nil
	}
	if name == "" {
		return fmt.Errorf("resource name is required")
	}
	return declarationNameRule(name)
}

func invokeDeclarationValidator(validator DeclarationValidator, name string) error {
	if validator == nil {
		return nil
	}
	return validator.Validate(name)
}
