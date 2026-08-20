package application

func declarationValidatorBuilder(enabled bool) DeclarationValidator {
	if !enabled {
		// Return a true nil interface so callers can detect the disabled
		// case with == nil. A typed nil pointer (var v *requiredNameValidator)
		// would carry a non-nil interface type, bypassing that check and
		// panicking when Validate is invoked.
		return nil
	}
	return &requiredNameValidator{}
}
