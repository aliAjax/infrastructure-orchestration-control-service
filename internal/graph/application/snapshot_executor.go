package application

import "sync"

// RunOnSnapshot executes the same resource snapshot for concurrent workers.
func RunOnSnapshot(resources []string, workers int, fn func(string)) {
	if workers < 1 {
		workers = 1
	}
	snapshot := append([]string(nil), resources...)
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
