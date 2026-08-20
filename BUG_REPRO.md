# Bug Reproduction

## Bug

A failed plan cannot transition back to executing, so retry requests remain stuck in the failed state.

## Trigger

Run:

```bash
go test ./internal/plan/application -run '^TestValidatePlanTransitionAllowsRetryAfterFailure$' -count=1
```

## Error

```text
--- FAIL: TestValidatePlanTransitionAllowsRetryAfterFailure
status_machine_test.go:11: failed plan should be retryable: invalid plan transition failed -> executing
```
