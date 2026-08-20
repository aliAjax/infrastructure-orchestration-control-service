# Bug Reproduction

## Bug

Audit persistence failures are converted to plain text and lose the original storage sentinel, so `errors.Is` cannot classify the failure.

## Trigger

Run:

```bash
go test ./internal/audit/application -run '^TestWrapAuditFailurePreservesSentinel$' -count=1
```

## Error

```text
--- FAIL: TestWrapAuditFailurePreservesSentinel
error_chain_test.go:12: sentinel was lost: audit persistence failed: storage unavailable
```
