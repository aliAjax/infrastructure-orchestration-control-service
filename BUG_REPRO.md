# Bug Reproduction

## Bug

Batch rollback ignores the first rollback error, continues to later snapshots, returns nil, and detaches the caller context.

## Trigger

Run:

```bash
go test ./internal/state/infrastructure -run '^TestRollbackEachStopsAtFirstErrorWithoutSkippingPriorWork$' -count=1
go test ./internal/state/infrastructure -run '^TestRollbackPolicyStopsOnError$' -count=1
go test ./internal/state/infrastructure -run '^TestRollbackPolicyKeepsContext$' -count=1
```

## Error

```text
rollback_batch_test.go:20: unexpected rollback flow calls=[s1 s2 s3] err=<nil>
rollback_policy_test.go:11: error was ignored
rollback_policy_test.go:17: context detached
```
