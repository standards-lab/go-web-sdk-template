package app_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-core/reactor"
	"github.com/standards-lab/go-web-sdk"

	"github.com/standards-lab/go-web-sdk-template/template/internal/app"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config/configtest"
)

// This file pins the graph's one structural property the composition root
// owns, the server alone in the top layer, over the root set App.Roots
// publishes, which it first holds to what a real Run builds. What that position buys at
// run time — startup in layer order, readiness held until every check
// passes, the server's runtime error watched as Monitored, and the drain in
// reverse layer order — is pinned in go-core's lifecycle tests and
// go-web-sdk's health and server tests, not here. Unlike app_test.go, these
// tests name nodes: they reach the graph through App.Graph and App.Nodes, as
// does the Build-failure case, which substitutes a failing constructor.

// buildAsRun builds a's graph from the set Run builds, App.Roots plus the
// lifecycle config node, plus extra. TestGraph_RootsBuildAsRun holds this
// set to Run's by observing a real Run.
func buildAsRun(t *testing.T, a *app.App, extra ...graph.Ref) *graph.System {
	t.Helper()
	roots := append(a.Roots(), a.Nodes().LifecycleConfig)
	sys, err := a.Graph().Build(append(roots, extra...)...)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return sys
}

// observeRun runs a with a context already ended, so the Run builds and
// starts nothing, and returns the handle of every node its Build
// constructed.
func observeRun(t *testing.T, a *app.App) []graph.Ref {
	t.Helper()
	var ran []graph.Ref
	a.Graph().Observe(func(r graph.Ref) { ran = append(ran, r) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if code := a.Run(ctx, io.Discard); code != 0 {
		t.Fatalf("Run = %d, want 0", code)
	}
	return ran
}

// A Build of App.Roots plus the lifecycle config node constructs exactly
// the nodes a real Run constructs, compared by handle: every node Run
// built is in the System, and the System holds no node Run did not build.
// It holds the tests below, which inspect buildAsRun, to what Run reaches.
func TestGraph_RootsBuildAsRun(t *testing.T) {
	a := app.New(configtest.Config(t), io.Discard)
	ran := observeRun(t, a)

	sys := buildAsRun(t, a)
	for _, r := range ran {
		if !sys.Has(r) {
			t.Errorf("Run built %q, the root set's Build did not", r.Name())
		}
	}
	if got, want := len(sys.Dependencies()), len(ran); got != want {
		t.Errorf("the root set's Build holds %d nodes, Run built %d; System:\n%s", got, want, sys)
	}
}

// serverAloneOnTop reports whether the System's last layer holds the server
// node alone, and returns that layer.
func serverAloneOnTop(sys *graph.System, n app.Nodes) ([]graph.Dependency, bool) {
	layers := sys.Layers()
	if len(layers) == 0 {
		return nil, false
	}
	top := layers[len(layers)-1]
	alone := len(top) == 1 &&
		top[0].Name == n.Server.Name() &&
		top[0].Value == any(sys.Get(n.Server))
	return top, alone
}

// The server is the only node in the top layer of the graph Run builds, so
// it starts after every other node and drains first. A node defined above
// it, or beside it at the top, fails here; the second case shows the check
// catches one.
func TestGraph_ServerAloneInTopLayer(t *testing.T) {
	t.Run("as defined", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		sys := buildAsRun(t, a)
		if top, ok := serverAloneOnTop(sys, a.Nodes()); !ok {
			t.Errorf("top layer = %+v, want the server node alone", names(top))
		}
	})

	t.Run("a node above the server", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		n := a.Nodes()
		above := a.Graph().Define("above", func(s *graph.Scope) (*web.Server, error) {
			return s.Use(n.Server), nil
		})
		sys := buildAsRun(t, a, above)
		if top, ok := serverAloneOnTop(sys, n); ok {
			t.Errorf("top layer = %+v, reported the server alone with a node above it", names(top))
		}
	})
}

// A reactor is reached by no Use, so Run builds it as a root, and the
// server orders itself after each reactor, so the server stays alone in the
// top layer above it. The template defines no reactor, so the fixture is
// added black-box, joining the reactors group with graph.In. It orders
// itself after the router, which puts it in the server's layer unless the
// server orders itself after it; a fixture lower in the graph would sit below the server
// regardless, and the test could not fail. Its interval is an hour, so it
// never ticks while the test runs.
func TestGraph_ServerAboveReactor(t *testing.T) {
	a := app.New(configtest.Config(t), io.Discard)
	n := a.Nodes()
	fixture := a.Graph().Define("fixture", func(*graph.Scope) (*reactor.Reactor[time.Time], error) {
		return reactor.New(reactor.Every(time.Hour), func(context.Context, time.Time) error {
			return nil
		}), nil
	}, graph.After(n.Router), graph.In(n.Reactors))

	// Run's own Build, observed, shows Run builds the reactor as a root:
	// no node Uses it, so nothing else would reach it. The context ends
	// before startup, so the Run starts nothing.
	if !slices.Contains(observeRun(t, a), graph.Ref(fixture)) {
		t.Fatalf("Run did not build the reactor %q", fixture.Name())
	}

	sys := buildAsRun(t, a)
	if top, ok := serverAloneOnTop(sys, a.Nodes()); !ok {
		t.Errorf("top layer = %+v, want the server node alone above the reactor", names(top))
	}
	if r, s := layerOf(sys, fixture.Name()), layerOf(sys, n.Server.Name()); r < 0 || r >= s {
		t.Errorf("reactor in layer %d, server in layer %d; want the reactor below the server", r, s)
	}
}

// layerOf returns the index of the layer holding the node named name, or -1.
func layerOf(sys *graph.System, name string) int {
	for i, layer := range sys.Layers() {
		for _, d := range layer {
			if d.Name == name {
				return i
			}
		}
	}
	return -1
}

// names lists the layer's node names.
func names(layer []graph.Dependency) []string {
	out := make([]string, len(layer))
	for i, d := range layer {
		out[i] = d.Name
	}
	return out
}

// A constructor's error fails Run's Build: Run returns 1 and writes the
// failure to stderr, labelled with the failing node's name. It logs
// nothing, since the node that failed may be the logger itself, and
// nothing reports ready.
func TestRun_BuildFailureExitsOne(t *testing.T) {
	var log, stderr syncBuffer
	a := app.New(configtest.Config(t), &log)
	a.Graph().Replace(a.Nodes().Logger, func(*graph.Scope) (*slog.Logger, error) {
		return nil, errors.New("logger unavailable")
	})

	// Cancelled up front, so a Build that wrongly succeeds drains at once
	// and fails the exit-code check instead of hanging.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if code := a.Run(ctx, &stderr); code != 1 {
		t.Errorf("Run = %d, want 1", code)
	}
	failure := regexp.MustCompile(`^service failed: .*logger.*logger unavailable`)
	if got := stderr.String(); !failure.MatchString(got) {
		t.Errorf("stderr = %q, want a service failure naming the logger node and its error", got)
	}
	if got := log.String(); got != "" {
		t.Errorf("log = %q, want nothing logged", got)
	}
}
