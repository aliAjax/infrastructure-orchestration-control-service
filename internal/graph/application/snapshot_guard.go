package application

// validSnapshotLength reports whether a resource snapshot is worth
// executing. A snapshot with no resources carries nothing to run, so the
// guard accepts a non-empty snapshot and rejects an empty one.
func validSnapshotLength(length int) bool { return length > 0 }
