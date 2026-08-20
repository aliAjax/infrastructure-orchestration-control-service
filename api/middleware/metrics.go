package middleware

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	requests atomic.Int64
	errors   atomic.Int64
	inFlight atomic.Int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) Track(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.requests.Add(1)
		m.inFlight.Add(1)
		defer m.inFlight.Add(-1)
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if sw.status >= 400 {
			m.errors.Add(1)
		}
	})
}

func (m *Metrics) Prometheus() []byte {
	return []byte(fmt.Sprintf(
		"# HELP http_requests_total Total HTTP requests.\n# TYPE http_requests_total counter\nhttp_requests_total %d\n# HELP http_errors_total Total HTTP errors.\n# TYPE http_errors_total counter\nhttp_errors_total %d\n# HELP http_in_flight Current in-flight requests.\n# TYPE http_in_flight gauge\nhttp_in_flight %d\n",
		m.requests.Load(), m.errors.Load(), m.inFlight.Load(),
	))
}
