# Bug Reproduction

## Bug

Drift-scan worker completion and error-stream lifecycle are inconsistent. Worker failures are dropped and the caller cannot reliably observe scan completion.

## Trigger

Run:

```bash
go test -race ./internal/drift/application -run '^TestScanEnvironmentsClosesErrorStreamAfterWorkers$' -count=1
go test -race ./internal/drift/application -run '^TestScanErrorPolicyCollectsFailures$' -count=1
go test -race ./internal/drift/application -run '^TestScanWorkerPolicyCountsAllEnvironments$' -count=1
```

## Error

```text
scan_lifecycle_test.go:37: unexpected errors=[]
scan_policy_test.go:10: scan failure was dropped
scan_policy_test.go:16: worker count mismatch
```
