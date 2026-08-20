package main

// Tests live in package main by exception to the black-box convention: a main
// package cannot be imported, so there is no external test package to use.

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
)

// testConfig builds a hermetic config: the log block finalizes directly with
// an empty prefix, so environment overrides stay disabled.
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	if err := cfg.Log.Finalize(""); err != nil {
		t.Fatalf("finalize log config: %v", err)
	}
	return cfg
}

// setInfrastructure is the cold start's first manifest: it must register the
// logger, wired to the writer it received, without touching the network.
func TestSetInfrastructure_RegistersTheLogger(t *testing.T) {
	var buf bytes.Buffer
	infra, err := setInfrastructure(&buf, testConfig(t))
	if err != nil {
		t.Fatalf("setInfrastructure: %v", err)
	}

	logger := infra.Get[*slog.Logger]()
	logger.Info("probe")
	if buf.Len() == 0 {
		t.Error("logger wrote nothing to the writer setInfrastructure received")
	}
}

func TestSetMiddleware_WiresTheRequestLogger(t *testing.T) {
	infra, err := setInfrastructure(&bytes.Buffer{}, testConfig(t))
	if err != nil {
		t.Fatalf("setInfrastructure: %v", err)
	}

	stack := setMiddleware(infra)
	if len(stack) != 1 || stack[0] == nil {
		t.Errorf("setMiddleware = %d entries, want the request logger alone", len(stack))
	}
}

// The baseline mounts no domain services; the manifest exists as the seam.
func TestSetRoutes_EmptyInTheBaseline(t *testing.T) {
	infra, err := setInfrastructure(&bytes.Buffer{}, testConfig(t))
	if err != nil {
		t.Fatalf("setInfrastructure: %v", err)
	}

	if modules := setRoutes(infra); len(modules) != 0 {
		t.Errorf("setRoutes = %d modules, want none in the baseline", len(modules))
	}
}
