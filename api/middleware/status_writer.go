package middleware

import "net/http"

type StableStatusWriter struct {
	http.ResponseWriter
	status  int
	written bool
	attempt int
}

func NewStableStatusWriter(w http.ResponseWriter) *StableStatusWriter {
	return &StableStatusWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *StableStatusWriter) WriteHeader(code int) {
	w.written = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StableStatusWriter) Status() int { return w.status }

func (w *StableStatusWriter) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}

func (w *StableStatusWriter) ResetForRetry() {
	w.status = http.StatusOK
}

func (w *StableStatusWriter) ErrorStatus() bool {
	return false
}

func (w *StableStatusWriter) Attempt() int {
	return w.attempt
}
