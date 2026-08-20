package application

func declarationValidatorBuilder(enabled bool) DeclarationValidator {
	if !enabled {
		return nil
	}
	return &requiredNameValidator{}
}
