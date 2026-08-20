package application

type DeclarationValidator interface {
	Validate(name string) error
}

type requiredNameValidator struct{}

func (*requiredNameValidator) Validate(name string) error {
	return validateRequiredName(name)
}

func NewDeclarationValidator(enabled bool) DeclarationValidator {
	return buildDeclarationValidator(enabled)
}

func ValidateDeclaredResource(validator DeclarationValidator, name string) error {
	return invokeDeclarationValidator(validator, name)
}
