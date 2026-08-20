package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStableStatusWriterKeepsFirstErrorStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewStableStatusWriter(recorder)
	writer.WriteHeader(http.StatusNotFound)
	writer.WriteHeader(http.StatusInternalServerError)
	if writer.Status() != http.StatusNotFound || recorder.Code != http.StatusNotFound {
		t.Fatalf("status was overwritten: wrapper=%d recorder=%d", writer.Status(), recorder.Code)
	}
}

func TestLoggingStatusWriterKeepsFirstHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &statusWriter{ResponseWriter: recorder, status: http.StatusOK}
	writer.WriteHeader(http.StatusUnauthorized)
	writer.WriteHeader(http.StatusInternalServerError)
	if writer.status != http.StatusUnauthorized {
		t.Fatalf("logging status changed after first header: %d", writer.status)
	}
}

func TestStableStatusWriterWriteCommitsOK(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewStableStatusWriter(recorder)
	if _, err := writer.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	writer.WriteHeader(http.StatusInternalServerError)
	if writer.Status() != http.StatusOK || recorder.Code != http.StatusOK {
		t.Fatalf("body write did not commit 200: wrapper=%d recorder=%d", writer.Status(), recorder.Code)
	}
}

func TestStableStatusWriterResetClearsWrittenState(t *testing.T) {
	writer := NewStableStatusWriter(httptest.NewRecorder())
	writer.WriteHeader(http.StatusServiceUnavailable)
	writer.ResetForRetry()
	if writer.written {
		t.Fatal("retry left the first-write latch set")
	}
	writer.WriteHeader(http.StatusAccepted)
	if writer.Status() != http.StatusAccepted {
		t.Fatalf("retry did not accept a new first status: %d", writer.Status())
	}
}

func TestStableStatusWriterTracksRetryAttempt(t *testing.T) {
	writer := NewStableStatusWriter(httptest.NewRecorder())
	writer.ResetForRetry()
	if writer.Attempt() != 2 {
		t.Fatalf("retry attempt was not advanced: %d", writer.Attempt())
	}
}

func TestStableStatusWriterReportsErrorClass(t *testing.T) {
	writer := NewStableStatusWriter(httptest.NewRecorder())
	writer.WriteHeader(http.StatusBadGateway)
	if !writer.ErrorStatus() {
		t.Fatal("5xx status was not classified as an error")
	}
}
