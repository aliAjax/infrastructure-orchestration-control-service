package application

import (
	"errors"
	"sync"
	"testing"
)

func TestScanEnvironmentsClosesErrorStreamAfterWorkers(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	errs := ScanEnvironments([]string{"prod", "stage", "dev"}, func(id string) error {
		mu.Lock()
		seen = append(seen, id)
		mu.Unlock()
		if id == "stage" {
			return errors.New("provider timeout")
		}
		return nil
	})
	if len(seen) != 3 || len(errs) != 1 || errs[0] == nil {
		t.Fatalf("workers did not complete cleanly seen=%v errs=%v", seen, errs)
	}
}
