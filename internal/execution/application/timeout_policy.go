package application

import "time"

func executionTimeout(value int64) time.Duration {
	if value <= 0 {
		return 30 * time.Second
	}
	return time.Duration(value) * time.Millisecond
}
