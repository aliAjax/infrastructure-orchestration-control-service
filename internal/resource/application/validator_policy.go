package application

func buildDeclarationValidator(enabled bool) DeclarationValidator {
	return declarationValidatorBuilder(enabled)
}

func validateRequiredName(_ string) error {
	if !declarationValidationGate() {
		return nil
	}
	return nil
}

func invokeDeclarationValidator(validator DeclarationValidator, name string) error {
	if validator == nil {
		return nil
	}
	return validator.Validate(name)
}
