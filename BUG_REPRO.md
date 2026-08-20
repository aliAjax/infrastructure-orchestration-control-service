# Bug Reproduction

## Bug

Cancellation is detached across the execution path. Work continues after the caller is cancelled, state updates use a detached context, and the configured task deadline is lost.

## Trigger

Run:

```bash
go test ./internal/execution/application -run '^TestRunWithRequestContextPropagatesCancellation$' -count=1
go test ./internal/execution/application -run '^TestExecutionContextRejectsCancelledInput$' -count=1
go test ./internal/execution/application -run '^TestRunnerContextIsCallerContext$' -count=1
go test ./internal/execution/application -run '^TestStateContextIsCallerContext$' -count=1
go test ./internal/execution/application -run '^TestExecutionTimeoutKeepsTaskDeadline$' -count=1
```

## Error

```text
context_guard_test.go:18: expected cancellation to reach work, called=true err=<nil>
context_policy_test.go:14: err=<nil>
context_policy_extra_test.go:11: runner context detached
context_policy_extra_test.go:17: state context detached
timeout_policy_test.go:10: task deadline lost
```
