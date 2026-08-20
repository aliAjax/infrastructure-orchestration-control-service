package application

import (
	"testing"
	"time"
)

func TestExecutionTimeoutKeepsTaskDeadline(t *testing.T) {
	if executionTimeout(1250) != 1250*time.Millisecond {
		t.Fatal("task deadline lost")
	}
}
