package app

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/standards-lab/go-web-sdk"
	mw "github.com/standards-lab/go-web-sdk/middleware"
)

// A handler's panic leaves the stack as a 500 problem that still carries
// the request id, and the log holds both the panic and the request record.
func TestMiddleware_RecoversPanic(t *testing.T) {
	var buf bytes.Buffer
	infra := &Infrastructure{Logger: slog.New(slog.NewTextHandler(&buf, nil))}
	handler := web.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), middleware(infra)...)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if rec.Header().Get(mw.RequestIDHeader) == "" {
		t.Errorf("response carries no %s header", mw.RequestIDHeader)
	}
	out := buf.String()
	if !strings.Contains(out, "boom") {
		t.Errorf("log has no panic record: %q", out)
	}
	if !strings.Contains(out, "http.response.status_code=500") {
		t.Errorf("log has no request record with the 500: %q", out)
	}
}
