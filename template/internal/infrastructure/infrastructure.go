package infrastructure

import (
	"context"
	"io"
	"log/slog"

	"github.com/standards-lab/go-core/logging"
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
)

// Infrastructure holds the shared subsystems every binary composes over —
// only the logger (inert — no lifecycle) in the baseline. Typed access flows
// through its fields; the composition root narrows them to primitives before
// they reach domain code.
type Infrastructure struct {
	Logger *slog.Logger
}

// New constructs every shared subsystem from the root configuration's
// blocks, with w receiving log output. Construction performs no I/O —
// connectivity waits for Start — and subsystem selection lives here, so
// every binary that constructs infrastructure inherits the same composition.
func New(w io.Writer, cfg *config.Config) (*Infrastructure, error) {
	return &Infrastructure{
		Logger: logging.New(w, cfg.Log),
	}, nil
}

// Start establishes connectivity for the lifecycle-bearing subsystems. The
// baseline has none, so it is a clean no-op; it carries the lifecycle hook
// signature, and a CLI without a coordinator calls it directly.
func (i *Infrastructure) Start(ctx context.Context) error {
	return nil
}

// Shutdown tears the shared subsystems down, in reverse dependency order as
// they arrive. The composition root runs it after the HTTP server has
// drained.
func (i *Infrastructure) Shutdown(ctx context.Context) error {
	return nil
}

// Checks yields the readiness checks of the subsystems that report
// readiness, derived from what the struct holds — a subsystem cannot be
// wired here yet missing from the probe.
func (i *Infrastructure) Checks() []web.Check {
	return nil
}
