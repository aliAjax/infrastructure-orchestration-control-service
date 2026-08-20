package middleware

import (
	"net/http"
	"sync"
	"time"
)

func RateLimit(limit int, window time.Duration, next http.Handler) http.Handler {
	if limit <= 0 {
		limit = 100
	}
	var mu sync.Mutex
	visitors := make(map[string][]time.Time)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		now := time.Now()
		mu.Lock()
		cutoff := now.Add(-window)
		times := visitors[key]
		idx := 0
		for idx < len(times) && times[idx].Before(cutoff) {
			idx++
		}
		times = append(times[idx:], now)
		if len(times) > limit {
			visitors[key] = times
			mu.Unlock()
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		visitors[key] = times
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
