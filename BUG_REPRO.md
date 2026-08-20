# Bug Reproduction

## Bug

The disabled declaration-validator path invokes a nil validator and panics, while the enabled path accepts an empty resource name.

## Trigger

Run:

```bash
go test ./internal/resource/application -run '^TestDisabledDeclarationValidatorDoesNotPanic$' -count=1
go test ./internal/resource/application -run '^TestEnabledDeclarationValidatorRejectsEmptyName$' -count=1
```

## Error

```text
panic: nil declaration validator
validator_test.go:13: enabled validator accepted an empty name
```
