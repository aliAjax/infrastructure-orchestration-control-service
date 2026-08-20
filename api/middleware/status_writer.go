package middleware

import "net/http"

type StableStatusWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func NewStableStatusWriter(w http.ResponseWriter) *StableStatusWriter {
	return &StableStatusWriter{ResponseWriter: w, status: http.StatusOK}
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
