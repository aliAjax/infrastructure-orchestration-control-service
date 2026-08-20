package application

type DeclarationValidator interface {
	Validate(name string) error
}

type requiredNameValidator struct{}

func (v *requiredNameValidator) Validate(name string) error {
	if v == nil {
		panic("nil declaration validator")
	}
	return validateRequiredName(name)
}

func NewDeclarationValidator(enabled bool) DeclarationValidator {
	return buildDeclarationValidator(enabled)
}

func ValidateDeclaredResource(validator DeclarationValidator, name string) error {
	return invokeDeclarationValidator(validator, name)
}
