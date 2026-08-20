package application

import "sync"

// RunOnSnapshot executes the same resource snapshot for concurrent workers.
// The caller's resources slice is treated as read-only input: workers run
// against an independent copy so concurrent invocations never write through
// to the caller's backing array and corrupt the names later layers see.
func RunOnSnapshot(resources []string, workers int, fn func(string)) {
	if workers < 1 {
		workers = 1
	}
	snapshot := make([]string, len(resources))
	copy(snapshot, resources)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, resource := range snapshot {
				fn(resource)
			}
		}()
	}
	wg.Wait()
}
