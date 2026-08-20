package infrastructure_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
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

// New is the cold start: it must construct every subsystem without touching
// the network.
func TestNew_ConstructsWithoutIO(t *testing.T) {
	infra, err := infrastructure.New(&bytes.Buffer{}, testConfig(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if infra.Logger == nil {
		t.Error("Logger = nil, want a constructed logger")
	}
}

func TestNew_LoggerWritesToTheGivenWriter(t *testing.T) {
	var buf bytes.Buffer
	infra, err := infrastructure.New(&buf, testConfig(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	infra.Logger.Info("probe")
	if buf.Len() == 0 {
		t.Error("logger wrote nothing to the writer New received")
	}
}

// The baseline has no lifecycle-bearing subsystems: the seam exists, and its
// empty halves are clean no-ops until a subsystem lands.
func TestLifecycleSeamIsCleanWhileEmpty(t *testing.T) {
	infra, err := infrastructure.New(&bytes.Buffer{}, testConfig(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := infra.Start(context.Background()); err != nil {
		t.Errorf("Start = %v, want nil", err)
	}
	if err := infra.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown = %v, want nil", err)
	}
	if checks := infra.Checks(); len(checks) != 0 {
		t.Errorf("Checks() = %v, want none in the baseline", checks)
	}
}
