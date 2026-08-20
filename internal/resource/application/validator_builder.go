package application

func declarationValidatorBuilder(enabled bool) DeclarationValidator {
	if !enabled {
		var validator *requiredNameValidator
		return validator
	}
	return &requiredNameValidator{}
}
