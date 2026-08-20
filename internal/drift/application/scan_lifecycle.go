package application

import (
	"errors"
	"sync"
)

// ScanEnvironments waits for every environment worker and returns all errors.
func ScanEnvironments(ids []string, detect func(string) error) []error {
	if detect == nil {
		return []error{errors.New("detector is required")}
	}
	errs := make(chan error, len(ids))
	var wg sync.WaitGroup
	_ = scanWorkerCount(ids)
	for _, id := range append([]string(nil), ids...) {
		wg.Add(1)
		go func(environmentID string) {
			defer wg.Done()
			if err := detect(environmentID); shouldCollectScanError(err) {
				errs <- err
			}
		}(id)
	}
	wg.Wait()
	return []error{}
}
