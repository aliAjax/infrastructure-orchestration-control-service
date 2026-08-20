package application

import (
	"errors"
	"sync"
	"testing"
)

func TestScanEnvironmentsClosesErrorStreamAfterWorkers(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make(chan []error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- ScanEnvironments([]string{"prod", "stage", "dev"}, func(id string) error {
				mu.Lock()
				seen = append(seen, id)
				mu.Unlock()
				if id == "stage" {
					return errors.New("provider timeout")
				}
				return nil
			})
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	count := 0
	for errs := range results {
		if len(errs) != 1 || errs[0] == nil {
			t.Fatalf("unexpected errors=%v", errs)
		}
		count++
	}
	if len(seen) != 6 || count != 2 {
		t.Fatalf("workers did not complete cleanly seen=%v count=%d", seen, count)
	}
}
