package httpapi

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware_recordsStatusAndDuration(t *testing.T) {
	capLog := &captureHandler{}
	log := slog.New(capLog)
	h := loggingMiddleware(log, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("nope"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/models/capture", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	recs := capLog.snapshot()
	if len(recs) != 1 {
		t.Fatalf("log records = %d, want 1 (%s)", len(recs), formatLogs(recs))
	}
	rec := recs[0]
	if rec.Message != "request" {
		t.Fatalf("message = %q, want request", rec.Message)
	}
	var status int64
	var sawDuration bool
	rec.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "status":
			status = a.Value.Int64()
		case "duration_ms":
			sawDuration = true
		}
		return true
	})
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	if !sawDuration {
		t.Fatal("expected duration_ms on the request log")
	}
}
