package middleware

import "net/http"

// StableStatusWriter pins the response to the first status committed via
// WriteHeader or Write. Once the response has been committed, any later
// WriteHeader is dropped so recovery and retry paths cannot overwrite the
// status the client already received. ResetForRetry clears the latch for a
// fresh attempt.
type StableStatusWriter struct {
	http.ResponseWriter
	status  int
	written bool
	attempt int
}

func NewStableStatusWriter(w http.ResponseWriter) *StableStatusWriter {
	return &StableStatusWriter{ResponseWriter: w, status: http.StatusOK, attempt: 1}
}

func (w *StableStatusWriter) WriteHeader(code int) {
	if w.written {
		// The response is already committed; keep the first status stable so
		// later code paths (recovery, retry) cannot change what the client saw.
		return
	}
	w.written = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StableStatusWriter) Status() int { return w.status }

func (w *StableStatusWriter) Write(p []byte) (int, error) {
	// A body write commits the implicit 200 OK just like net/http does.
	w.written = true
	return w.ResponseWriter.Write(p)
}

func (w *StableStatusWriter) ResetForRetry() {
	w.status = http.StatusOK
	w.written = false
	w.attempt++
}

func (w *StableStatusWriter) ErrorStatus() bool {
	return w.status >= 400
}

func (w *StableStatusWriter) Attempt() int {
	return w.attempt
}
