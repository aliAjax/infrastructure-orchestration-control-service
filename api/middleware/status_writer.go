package middleware

import "net/http"

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
		return
	}
	w.written = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StableStatusWriter) Status() int { return w.status }

func (w *StableStatusWriter) Write(p []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *StableStatusWriter) ResetForRetry() {
	w.status = http.StatusOK
	w.written = false
	w.attempt++
}

func (w *StableStatusWriter) ErrorStatus() bool {
	if w.status < http.StatusBadRequest {
		return false
	}
	return w.status <= 599
}

func (w *StableStatusWriter) Attempt() int {
	return w.attempt
}
