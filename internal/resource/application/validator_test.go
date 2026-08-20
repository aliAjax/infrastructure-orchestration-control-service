package application

import "testing"

func TestDisabledDeclarationValidatorDoesNotPanic(t *testing.T) {
	if err := ValidateDeclaredResource(NewDeclarationValidator(false), ""); err != nil {
		t.Fatalf("disabled validator should be a no-op: %v", err)
	}
}

func TestEnabledDeclarationValidatorRejectsEmptyName(t *testing.T) {
	if err := ValidateDeclaredResource(NewDeclarationValidator(true), ""); err == nil {
		t.Fatal("enabled validator accepted an empty name")
	}
}
