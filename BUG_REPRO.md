# Bug Reproduction

## Bug

Concurrent graph execution mutates the caller's resource slice and retains nested aliases instead of operating on an isolated snapshot.

## Trigger

Run:

```bash
go test -race ./internal/graph/application -run '^TestRunOnSnapshotKeepsInputStableDuringWorkers$' -count=1
go test -race ./internal/graph/application -run '^TestCloneResourceSnapshotDeepCopiesNestedState$' -count=1
go test -race ./internal/graph/application -run '^TestSnapshotCapacityPreservesResourceLength$' -count=1
go test -race ./internal/graph/application -run '^TestSnapshotLengthGuardAcceptsResourceCount$' -count=1
```

## Error

```text
WARNING: DATA RACE
resource shared:shared:db seen 4 times
snapshot retained nested aliases
snapshot capacity=0
valid snapshot length rejected
```
