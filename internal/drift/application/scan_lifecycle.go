package application

import (
	"errors"
	"sync"
)

// ScanEnvironments runs detect for every environment concurrently and returns
// every collected failure. It is safe to call from multiple goroutines.
func ScanEnvironments(ids []string, detect func(string) error) []error {
	if detect == nil {
		return []error{errors.New("detector is required")}
	}
	workerCount := scanWorkerCount(ids)
	errs := make(chan error, workerCount)
	var wg sync.WaitGroup
	for _, environmentID := range append([]string(nil), ids...) {
		wg.Add(1)
		go func(environmentID string) {
			defer wg.Done()
			if err := detect(environmentID); shouldCollectScanError(err) {
				errs <- err
			}
		}(environmentID)
	}
	wg.Wait()
	close(errs)
	collected := make([]error, 0, workerCount)
	for err := range errs {
		collected = append(collected, err)
	}
	return collected
}
