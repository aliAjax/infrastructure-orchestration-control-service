package application

import "time"

func executionTimeout(value int64) time.Duration { return 30 * time.Second }
