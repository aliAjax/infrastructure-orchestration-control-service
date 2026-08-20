package application

import "sync"

// ScanEnvironments waits for every environment worker and returns all errors.
func ScanEnvironments(ids []string, detect func(string) error) []error {
	errs := make(chan error, len(ids))
	var wg sync.WaitGroup
	for _, id := range append([]string(nil), ids...) {
		wg.Add(1)
		go func(environmentID string) {
			defer wg.Done()
			if err := detect(environmentID); err != nil {
				errs <- err
			}
		}(id)
	}
	wg.Wait()
	close(errs)
	var out []error
	for err := range errs {
		out = append(out, err)
	}
	return out
}
