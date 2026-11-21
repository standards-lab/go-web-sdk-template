package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	libconfig "github.com/standards-lab/go-core/config"
	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-web-sdk"
	mw "github.com/standards-lab/go-web-sdk/middleware"

	"github.com/standards-lab/go-web-sdk-template/template/internal/app"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config/configtest"
)

const failsafe = 2 * time.Second

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

var readyRecord = regexp.MustCompile(`msg="server ready" addr=(\S+)`)

func waitForReady(t *testing.T, buf *syncBuffer) string {
	t.Helper()
	deadline := time.Now().Add(failsafe)
	for time.Now().Before(deadline) {
		if m := readyRecord.FindStringSubmatch(buf.String()); m != nil {
			return m[1]
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for the server to become ready; log: %q", buf.String())
	return ""
}

type running struct {
	log    *syncBuffer
	stderr *syncBuffer
	cancel context.CancelFunc
	done   chan int
}

func start(t *testing.T, cfg *config.Config) *running {
	t.Helper()
	buf := &syncBuffer{}
	a := app.New(cfg, buf)
	ctx, cancel := context.WithCancel(context.Background())
	r := &running{log: buf, stderr: &syncBuffer{}, cancel: cancel, done: make(chan int, 1)}
	go func() { r.done <- a.Run(ctx, r.stderr) }()
	t.Cleanup(cancel)
	return r
}

func (r *running) exit(t *testing.T) int {
	t.Helper()
	select {
	case code := <-r.done:
		return code
	case <-time.After(failsafe):
		t.Fatalf("timed out waiting for Run to return; log: %q, stderr: %q", r.log.String(), r.stderr.String())
		return -1
	}
}

var client = &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

func get(t *testing.T, addr, path string) (int, http.Header, string) {
	t.Helper()
	resp, err := client.Get(fmt.Sprintf("http://%s%s", addr, path))
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s body: %v", path, err)
	}
	return resp.StatusCode, resp.Header, string(body)
}

func assertProbeHeaders(t *testing.T, path string, header http.Header) {
	t.Helper()
	if got := header.Get("Content-Type"); got != "application/json" {
		t.Errorf("GET %s Content-Type = %q, want application/json", path, got)
	}
	if got := header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("GET %s Cache-Control = %q, want no-store", path, got)
	}
	if header.Get(mw.RequestIDHeader) == "" {
		t.Errorf("GET %s carries no %s response header; RequestID is not wired", path, mw.RequestIDHeader)
	}
}

func TestRun_ServesProbesThenDrains(t *testing.T) {
	r := start(t, configtest.Config(t))
	addr := waitForReady(t, r.log)

	if code, _, _ := get(t, addr, web.HealthPath); code != http.StatusOK {
		t.Errorf("GET %s = %d, want 200", web.HealthPath, code)
	}
	if code, _, _ := get(t, addr, web.ReadyPath); code != http.StatusOK {
		t.Errorf("GET %s = %d, want 200", web.ReadyPath, code)
	}

	r.cancel()
	if code := r.exit(t); code != 0 {
		t.Errorf("Run = %d, want 0; log: %q", code, r.log.String())
	}

	out := r.log.String()
	if !strings.Contains(out, `msg="server stopped"`) {
		t.Error("log carries no stop record after the drain")
	}
	if !strings.Contains(out, "url.path="+web.HealthPath) {
		t.Error("log has no probe request record; the middleware stack is not wired")
	}
}

func TestRun_HealthzContract(t *testing.T) {
	r := start(t, configtest.Config(t))
	addr := waitForReady(t, r.log)

	code, header, body := get(t, addr, web.HealthPath)
	if code != http.StatusOK {
		t.Errorf("GET %s = %d, want 200", web.HealthPath, code)
	}
	assertProbeHeaders(t, web.HealthPath, header)
	if got := strings.TrimSpace(body); got != `{"status":"ok"}` {
		t.Errorf("GET %s body = %q, want {\"status\":\"ok\"}", web.HealthPath, got)
	}
}

func TestRun_ReadyzContract(t *testing.T) {
	r := start(t, configtest.Config(t))
	addr := waitForReady(t, r.log)

	code, header, body := get(t, addr, web.ReadyPath)
	if code != http.StatusOK {
		t.Errorf("GET %s = %d, want 200", web.ReadyPath, code)
	}
	assertProbeHeaders(t, web.ReadyPath, header)

	type check struct {
		Name  string `json:"name"`
		Ready bool   `json:"ready"`
	}
	var got struct {
		Status string  `json:"status"`
		Checks []check `json:"checks"`
	}
	dec := json.NewDecoder(strings.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("decode %s body %q: %v", web.ReadyPath, body, err)
	}
	if got.Status != "ready" {
		t.Errorf("readiness status = %q, want ready", got.Status)
	}
	want := []check{{Name: "lifecycle", Ready: true}}
	if len(got.Checks) != len(want) || got.Checks[0] != want[0] {
		t.Errorf("readiness checks = %+v, want exactly %+v", got.Checks, want)
	}
}

func TestRun_ReadyRecordNamesBoundAddr(t *testing.T) {
	r := start(t, configtest.Config(t))
	addr := waitForReady(t, r.log)

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("ready addr %q: %v", addr, err)
	}
	if host != "127.0.0.1" {
		t.Errorf("ready addr host = %q, want the configured 127.0.0.1", host)
	}
	if p, err := strconv.Atoi(port); err != nil || p == 0 {
		t.Errorf("ready addr port = %q, want the bound nonzero port", port)
	}
	if code, _, _ := get(t, addr, web.ReadyPath); code != http.StatusOK {
		t.Errorf("GET %s at the logged addr = %d, want 200", web.ReadyPath, code)
	}
	if n := len(readyRecord.FindAllString(r.log.String(), -1)); n != 1 {
		t.Errorf("ready records = %d, want 1", n)
	}
}

func TestRun_TakenPortFailsStartup(t *testing.T) {
	held, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("hold a port: %v", err)
	}
	defer func() { _ = held.Close() }()

	cfg := configtest.Config(t)
	cfg.Server.Port = new(held.Addr().(*net.TCPAddr).Port)

	r := start(t, cfg)
	if code := r.exit(t); code != 1 {
		t.Errorf("Run = %d, want 1", code)
	}

	out := r.log.String()
	failure := regexp.MustCompile(`msg="service failed" error="[^"]*startup:[^"]*server`)
	if !failure.MatchString(out) {
		t.Errorf("log = %q, want a startup error naming server", out)
	}
	if strings.Contains(out, "server ready") {
		t.Errorf("log = %q, want no ready record", out)
	}
}

func TestRun_ShutdownBoundedByTimeout(t *testing.T) {
	const timeout = 200 * time.Millisecond
	const margin = 500 * time.Millisecond

	cfg := configtest.Config(t)
	cfg.ShutdownTimeout = libconfig.Duration(timeout)

	r := start(t, cfg)
	addr := waitForReady(t, r.log)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial %s: %v", addr, err)
	}
	defer func() { _ = conn.Close() }()
	unterminatedHeaders := "GET " + web.HealthPath + " HTTP/1.1\r\nHost: stall\r\n"
	if _, err := io.WriteString(conn, unterminatedHeaders); err != nil {
		t.Fatalf("write partial request: %v", err)
	}
	const settleBeforeDrain = 50 * time.Millisecond
	time.Sleep(settleBeforeDrain)

	began := time.Now()
	r.cancel()
	code := r.exit(t)
	elapsed := time.Since(began)

	if code != 1 {
		t.Errorf("Run = %d, want 1 when the drain outlives the timeout; log: %q", code, r.log.String())
	}
	if elapsed > timeout+margin {
		t.Errorf("Run returned after %s, want within %s of the %s timeout", elapsed, margin, timeout)
	}
}

func TestRun_TwicePanics(t *testing.T) {
	a := app.New(configtest.Config(t), io.Discard)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if code := a.Run(ctx, io.Discard); code != 0 {
		t.Fatalf("first Run = %d, want 0", code)
	}

	built := 0
	a.Graph().Observe(func(graph.Ref) { built++ })
	defer func() {
		if recover() == nil {
			t.Error("a second Run did not panic")
		}
		if built != 0 {
			t.Errorf("the second Run built %d nodes before panicking, want none", built)
		}
	}()
	_ = a.Run(ctx, io.Discard)
}
