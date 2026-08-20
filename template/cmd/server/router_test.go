package main

// Tests live in package main by exception to the black-box convention: a main
// package cannot be imported, so there is no external test package to use.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

// failsafe bounds every wait for an event that should occur, so a broken
// coordinator fails the test instead of hanging it.
const failsafe = 2 * time.Second

func probe(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func testInfra(t *testing.T) *infrastructure.Infrastructure {
	t.Helper()
	cfg := &config.Config{}
	if err := cfg.Log.Finalize(""); err != nil {
		t.Fatalf("finalize log config: %v", err)
	}
	infra, err := infrastructure.New(io.Discard, cfg)
	if err != nil {
		t.Fatalf("construct infrastructure: %v", err)
	}
	return infra
}

func TestRouter_Healthz(t *testing.T) {
	h := newRouter(testInfra(t), lifecycle.New())

	rec := probe(t, h, http.MethodGet, web.HealthPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal %q: %v", rec.Body.String(), err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
}

func TestRouter_ReadyzNotReadyBeforeStartup(t *testing.T) {
	h := newRouter(testInfra(t), lifecycle.New())

	if got := probe(t, h, http.MethodGet, web.ReadyPath).Code; got != http.StatusServiceUnavailable {
		t.Errorf("status = %d before startup, want 503", got)
	}
}

func TestRouter_ReadyzReadyAfterStartup(t *testing.T) {
	lc := lifecycle.New()
	h := newRouter(testInfra(t), lc)

	ready := make(chan struct{})
	lc.OnReady(func() { close(ready) })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- lc.Run(ctx, time.Second) }()

	select {
	case <-ready:
	case <-time.After(failsafe):
		t.Fatal("timed out waiting for the coordinator to become ready")
	}

	if got := probe(t, h, http.MethodGet, web.ReadyPath).Code; got != http.StatusOK {
		t.Errorf("status = %d after startup, want 200", got)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(failsafe):
		t.Fatal("timed out waiting for Run to return")
	}
}

func TestRouter_ProbesRejectOtherMethods(t *testing.T) {
	h := newRouter(testInfra(t), lifecycle.New())

	if got := probe(t, h, http.MethodPost, web.HealthPath).Code; got != http.StatusMethodNotAllowed {
		t.Errorf("POST %s = %d, want 405", web.HealthPath, got)
	}
}
