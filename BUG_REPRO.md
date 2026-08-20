# Bug Reproduction

## What is wrong

Lock renewal and release do not have a complete lifecycle. Renewal can run after release starts, cleanup errors are discarded, cancelled callers lose cleanup values, and invalid lock inputs reach the manager.

## How to trigger

Run the lock application lifecycle tests with a non-positive TTL, a nil acquired lock, overlapping renewal/release, cancellation during cleanup, renewal failure, release failure, and simultaneous operation and release failures.

## Observed error

The invalid TTL path can panic with:

```text
panic: non-positive interval for NewTicker

goroutine 4 [running]:
time.NewTicker(0x0)
    /Users/zhuanzmima0000/.local/go/src/time/tick.go:22 +0x16c
github.com/infra-orchestration/controlplane/internal/lock/application.(*Service).renew(...)
    /app/internal/lock/application/service.go:35 +0x48
```

Other baseline failures include `release raced renewal exit`, `release error was lost`, and `cleanup did not preserve both err`.
