package app_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/app"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

// failsafe bounds every wait for an event that should occur, so a broken
// coordinator fails the test instead of hanging it.
const failsafe = 2 * time.Second

// checker is a fixed-answer readiness check for registrations under test.
type checker bool

func (c checker) Ready() bool { return bool(c) }

// syncBuffer serializes writes so the app's logging goroutines and the
// test's reads stay race-free.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// testConfig builds a hermetic config: loopback host, an explicit zero port
// for an ephemeral listener, and an empty prefix so environment overrides
// stay disabled.
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = new(int)
	if err := cfg.Finalize(""); err != nil {
		t.Fatalf("finalize config: %v", err)
	}
	return cfg
}

// testRegistry builds a registry holding the logger every App requires,
// writing to the returned buffer, plus any extra services.
func testRegistry(t *testing.T, extra ...infrastructure.Service) (*infrastructure.Registry, *syncBuffer) {
	t.Helper()
	buf := &syncBuffer{}
	r := infrastructure.NewRegistry()
	r.Register(slog.New(slog.NewTextHandler(buf, nil)), infrastructure.Service{Name: "logger"})
	for i, svc := range extra {
		r.Register(i, svc)
	}
	return r, buf
}

// waitForReady polls the log for the coordinator's ready record and returns
// the address the server bound.
func waitForReady(t *testing.T, buf *syncBuffer) string {
	t.Helper()
	deadline := time.Now().Add(failsafe)
	for time.Now().Before(deadline) {
		if out := buf.String(); strings.Contains(out, "server ready") {
			_, after, ok := strings.Cut(out, "addr=")
			if !ok {
				t.Fatalf("ready record carries no addr: %q", out)
			}
			return strings.Fields(after)[0]
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for the server to become ready; log: %q", buf.String())
	return ""
}

func get(t *testing.T, addr, path string) int {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("http://%s%s", addr, path))
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestRun_ServesProbesAndMountedModulesThenDrains(t *testing.T) {
	infra, buf := testRegistry(t, infrastructure.Service{Name: "extra", Check: checker(true)})

	group := web.NewGroup("/t")
	group.HandleFunc(http.MethodGet, "/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	a := app.New(testConfig(t), infra, app.Wiring{
		Modules: []*web.Module{web.NewModule(group)},
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() { done <- a.Run(ctx) }()

	addr := waitForReady(t, buf)
	if code := get(t, addr, web.HealthPath); code != http.StatusOK {
		t.Errorf("GET %s = %d, want 200", web.HealthPath, code)
	}
	if code := get(t, addr, web.ReadyPath); code != http.StatusOK {
		t.Errorf("GET %s = %d, want 200", web.ReadyPath, code)
	}
	if code := get(t, addr, "/t/ping"); code != http.StatusNoContent {
		t.Errorf("GET /t/ping = %d, want 204", code)
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Errorf("Run = %d, want 0", code)
		}
	case <-time.After(failsafe):
		t.Fatal("timed out waiting for Run to return")
	}
	if !strings.Contains(buf.String(), "server stopped") {
		t.Error("log carries no stop record after the drain")
	}
}

// The registry's checks feed the probe: a service that reports unready keeps
// /readyz at 503 even after the coordinator's startup completes.
func TestRun_ReadyzReportsAnUnreadyService(t *testing.T) {
	infra, buf := testRegistry(t, infrastructure.Service{Name: "warming", Check: checker(false)})

	a := app.New(testConfig(t), infra, app.Wiring{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() { done <- a.Run(ctx) }()

	addr := waitForReady(t, buf)
	if code := get(t, addr, web.ReadyPath); code != http.StatusServiceUnavailable {
		t.Errorf("GET %s = %d with an unready service, want 503", web.ReadyPath, code)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(failsafe):
		t.Fatal("timed out waiting for Run to return")
	}
}

// Every App requires the logger; assembling one from a registry without it
// is a wiring mistake and panics at cold start.
func TestNew_PanicsWithoutALogger(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("New with no registered logger did not panic")
		}
	}()
	app.New(testConfig(t), infrastructure.NewRegistry(), app.Wiring{})
}
