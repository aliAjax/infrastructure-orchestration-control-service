package application

import "time"

// executionTimeout preserves the task's own deadline rather than overwriting it
// with a fixed value, so a per-task timeout is honoured exactly as configured.
func executionTimeout(value int64) time.Duration { return time.Duration(value) * time.Millisecond }
