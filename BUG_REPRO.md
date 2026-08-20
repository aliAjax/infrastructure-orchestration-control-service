# Bug Reproduction

## Bug

Resource filtering and listing expose shared slice and nested storage. Mutating a returned resource can pollute the original input or repository-owned result.

## Trigger

Run:

```bash
go test ./internal/resource/application -run '^TestFilterResourcesDoesNotPolluteOriginalSlice$' -count=1
go test ./internal/resource/application -run '^TestFilterResourcesReturnsConcreteEmptySlice$' -count=1
go test ./internal/resource/application -run '^TestFilterResourcesCopiesDependencyIDs$' -count=1
go test ./internal/resource/application -run '^TestFilterResourcesCopiesDesiredState$' -count=1
go test ./internal/resource/application -run '^TestListResourcesDetachesRepositoryResult$' -count=1
```

## Error

```text
resource_filter_test.go:17: original resources were polluted
resource_filter_test.go:24: empty filter result must be non-nil: []domain.Resource(nil)
resource_filter_test.go:33: dependency IDs share storage: []string{"mutated"}
resource_filter_test.go:42: desired state shares storage
resource_filter_test.go:62: environment identity was not normalized
```
