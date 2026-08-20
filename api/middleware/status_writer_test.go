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
