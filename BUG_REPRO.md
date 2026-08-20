# Bug Reproduction

## Bug

Configuration loading does not honor an already-cancelled caller context. The load callback still runs and no cancellation error is returned.

## Trigger

Run:

```bash
go test ./internal/platform -run '^TestConfigLoaderDoesNotReuseCancelledContext$' -count=1
```

## Error

```text
--- FAIL: TestConfigLoaderDoesNotReuseCancelledContext
config_lifecycle_test.go:18: cancelled context was not honored called=true err=<nil>
```
