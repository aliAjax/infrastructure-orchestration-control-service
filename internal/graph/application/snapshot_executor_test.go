package application

import (
	"sync"
	"testing"
)

func TestRunOnSnapshotKeepsInputStableDuringWorkers(t *testing.T) {
	resources := []string{"db", "cache", "queue"}
	var mu sync.Mutex
	seen := make(map[string]int)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			RunOnSnapshot(resources, 4, func(resource string) {
				mu.Lock()
				seen[resource]++
				mu.Unlock()
			})
		}()
	}
	close(start)
	wg.Wait()
	for _, resource := range resources {
		if seen[resource] != 8 {
			t.Fatalf("resource %s seen %d times", resource, seen[resource])
		}
	}
	if resources[0] != "db" || resources[1] != "cache" || resources[2] != "queue" {
		t.Fatalf("input snapshot was mutated: %#v", resources)
	}
}
