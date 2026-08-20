package application

// snapshotCapacity returns the capacity to reserve for a snapshot holding
// the given number of resources, so the clone is not silently shrunk or
// re-grown mid-run. It mirrors the resource count directly: a snapshot must
// preserve every resource in the input without dropping or padding it.
func snapshotCapacity(length int) int {
	if length < 0 {
		return 0
	}
	return length
}
