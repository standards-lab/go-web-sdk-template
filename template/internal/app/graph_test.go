package app_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-core/graph/graphtest"
	"github.com/standards-lab/go-core/reactor"
	"github.com/standards-lab/go-web-sdk"

	"github.com/standards-lab/go-web-sdk-template/template/internal/app"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config/configtest"
)

// This file pins the graph's structural properties the composition root
// owns: each Nodes field names its node, the root set App.Roots publishes
// builds what a real Run builds, the server is alone in the top layer over
// it, a reactor sits below the server by its group membership alone, and
// the root set drops no order but the one an empty reactors group leaves.
// What that position buys at run time — startup in layer order, readiness
// held until every check passes, the server's runtime error watched as
// Monitored, and the drain in reverse layer order — is pinned in go-core's
// lifecycle tests and go-web-sdk's health and server tests, not here.
// Unlike app_test.go, these tests reach the graph through App.Graph and
// App.Nodes, and they read placement by handle: the System's layer of each
// node handle, and the handle on each built node.
//
// Each placement guard is a function returning an error that carries the
// System's layer table, called both by the test of the App's own roots and
// by a test of a fixture that breaks it, so the fixture shows the real
// assertion can fail.

// Each node's name is its Nodes field's name lowercased, ignoring the
// name's spaces, so a field names the node it holds without a comment
// repeating the name, and every field holds a handle once New returns.
func TestGraph_NodesNamedByTheirFields(t *testing.T) {
	a := app.New(configtest.Config(t), io.Discard)
	graphtest.CheckNames(t, a.Nodes())
}

// runSet returns the set Run builds, App.Roots plus the lifecycle config
// node, plus extra. TestGraph_RootsBuildAsRun holds this set to Run's by
// observing a real Run.
func runSet(a *app.App, extra ...graph.Ref) []graph.Ref {
	return append(append(a.Roots(), a.Nodes().LifecycleConfig), extra...)
}

// buildAsRun builds a's graph from runSet(a, extra...).
func buildAsRun(t *testing.T, a *app.App, extra ...graph.Ref) *graph.System {
	t.Helper()
	sys, err := a.Graph().Build(runSet(a, extra...)...)
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

// layerError returns an error of the formatted message followed by sys's
// layer table.
func layerError(sys *graph.System, format string, args ...any) error {
	return fmt.Errorf(format+"\nlayers:\n%s", append(args, sys)...)
}

// wantGuardFails fails t unless err, a guard's verdict on sys, is an error
// that names mention and carries sys's layer table.
func wantGuardFails(t *testing.T, err error, sys *graph.System, mention string) {
	t.Helper()
	switch {
	case err == nil:
		t.Fatalf("the guard passed a fixture that breaks it\nlayers:\n%s", sys)
	case !strings.Contains(err.Error(), mention):
		t.Errorf("guard error = %q, want it to name %s", err, mention)
	case !strings.Contains(err.Error(), "\nlayers:\n"+sys.String()):
		t.Errorf("guard error = %q, want it to carry the layer table:\n%s", err, sys)
	}
}

// A Build of App.Roots plus the lifecycle config node constructs exactly
// the nodes a real Run constructs, compared by handle: every node Run
// built is in the System, and the System holds no node Run did not build.
// It holds the tests below, which inspect that set, to what Run reaches.
func TestGraph_RootsBuildAsRun(t *testing.T) {
	a := app.New(configtest.Config(t), io.Discard)
	ran := observeRun(t, a)

	sys := buildAsRun(t, a)
	for _, r := range ran {
		if !sys.Has(r) {
			t.Errorf("Run built %s, the root set's Build did not", r.Name())
		}
	}
	for _, d := range sys.Dependencies() {
		if !slices.Contains(ran, d.Ref) {
			t.Errorf("the root set's Build built %s, Run did not", d.Name)
		}
	}
	if t.Failed() {
		t.Logf("layers:\n%s", sys)
	}
}

// serverAlone reports, as an error carrying the layer table, any node of
// sys at or above the server's layer: the server must be alone in the top
// layer, so it starts after every other node and drains first.
func serverAlone(sys *graph.System, server graph.Ref) error {
	if !sys.Has(server) {
		return layerError(sys, "%s is not in the System", server.Name())
	}
	top := sys.LayerOf(server)
	var beside []string
	for _, d := range sys.Dependencies() {
		if d.Ref != server && sys.LayerOf(d.Ref) >= top {
			beside = append(beside, d.Name)
		}
	}
	if len(beside) > 0 {
		return layerError(sys, "%v at or above %s's layer %d; want %s alone in the top layer",
			beside, server.Name(), top, server.Name())
	}
	return nil
}

// The server is the only node in the top layer of the graph Run builds. A
// node defined above it, or beside it at the top, fails the guard; the
// second case shows the guard catches one.
func TestGraph_ServerAloneInTopLayer(t *testing.T) {
	t.Run("as defined", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		if err := serverAlone(buildAsRun(t, a), a.Nodes().Server); err != nil {
			t.Error(err)
		}
	})

	t.Run("a node above the server", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		n := a.Nodes()
		above := a.Graph().Define("above", func(s *graph.Scope) (*web.Server, error) {
			return s.Use(n.Server), nil
		})
		sys := buildAsRun(t, a, above)
		wantGuardFails(t, serverAlone(sys, n.Server), sys, "above")
	})
}

// fixtureReactor defines on a's graph a reactor named name, ordered after
// the router with opts besides. Ordered after the router, it sits in the
// server's layer unless the server orders itself after it; a reactor lower
// in the graph would sit below the server regardless, and the guard could
// not fail. Its interval is an hour, so it never ticks while a test runs.
func fixtureReactor(a *app.App, name string, opts ...graph.Option) *graph.Node[*reactor.Reactor[time.Time]] {
	opts = append([]graph.Option{graph.After(a.Nodes().Router)}, opts...)
	return a.Graph().Define(name, func(*graph.Scope) (*reactor.Reactor[time.Time], error) {
		return reactor.New(reactor.Every(time.Hour), func(context.Context, time.Time) error {
			return nil
		}), nil
	}, opts...)
}

// A reactor joins the reactors group with graph.In on its Define line and
// declares nothing else: Run builds it, since the group is one of the
// roots, and it sits below the server, since the server orders itself
// after the group. The template defines no reactor, so the fixture is
// added black-box through the group. The second case shows the guard
// catches the same reactor built without its In: nothing orders the server
// after it, so it sits beside the server at the top.
func TestGraph_ServerAboveReactor(t *testing.T) {
	t.Run("in the reactors group", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		fixture := fixtureReactor(a, "fixture", graph.In(a.Nodes().Reactors))

		// Run's own Build, observed, shows Run builds the reactor: no node
		// Uses it, so only the group's place in the roots reaches it.
		if !slices.Contains(observeRun(t, a), graph.Ref(fixture)) {
			t.Fatalf("Run did not build the reactor %s", fixture.Name())
		}
		if err := serverAlone(buildAsRun(t, a), a.Nodes().Server); err != nil {
			t.Error(err)
		}
	})

	t.Run("outside the reactors group", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		stray := fixtureReactor(a, "stray")
		sys := buildAsRun(t, a, stray)
		wantGuardFails(t, serverAlone(sys, a.Nodes().Server), sys, "stray")
	})
}

// rootDrops builds roots from a's graph, a set holding App.Roots, and
// reports, as an error carrying the layer table, a failed Build or any
// After the Build dropped beyond the one allowed: while the reactors group
// has no member, the server's After of the group orders nothing and is
// dropped, so exactly that Drop is reported; with any member, nothing is.
// Any other drop means a root is missing from the set, or a reactor its
// In, so an order is silently lost. Whether the group has a member is read
// from a Build of the group alone, which constructs nothing when it has
// none.
func rootDrops(a *app.App, roots []graph.Ref) error {
	n := a.Nodes()
	members, err := a.Graph().Build(n.Reactors)
	if err != nil {
		return fmt.Errorf("Build of %s: %w", n.Reactors.Name(), err)
	}
	empty := len(members.Dependencies()) == 0

	sys, err := a.Graph().Build(roots...)
	if err != nil {
		return fmt.Errorf("Build: %w", err)
	}
	allowed := graph.Drop{Node: n.Server, After: n.Reactors}
	var drops []string
	sawAllowed := false
	for _, d := range sys.Dropped() {
		if empty && d == allowed {
			sawAllowed = true
			continue
		}
		drops = append(drops, fmt.Sprintf("%s's After of %s", d.Node.Name(), d.After.Name()))
	}
	if len(drops) > 0 {
		return layerError(sys, "dropped %s; want every After to order", strings.Join(drops, ", "))
	}
	if empty && !sawAllowed {
		return layerError(sys, "%s's After of the empty %s group was not dropped; want exactly that Drop",
			n.Server.Name(), n.Reactors.Name())
	}
	return nil
}

// A Build of the web root set drops no After but the server's of the
// reactors group, and that one only while the group is empty. The third
// case shows the guard catches an After whose target the roots do not
// build.
func TestGraph_RootsDropOnlyTheEmptyReactors(t *testing.T) {
	t.Run("as defined", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		if err := rootDrops(a, runSet(a)); err != nil {
			t.Error(err)
		}
	})

	t.Run("with a reactor", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		fixtureReactor(a, "fixture", graph.In(a.Nodes().Reactors))
		if err := rootDrops(a, runSet(a)); err != nil {
			t.Error(err)
		}
	})

	t.Run("an After of a node not built", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		unbuilt := a.Graph().Value("unbuilt", 0)
		orderer := a.Graph().Value("orderer", 0, graph.After(unbuilt))
		roots := runSet(a, orderer)
		sys, err := a.Graph().Build(roots...)
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		wantGuardFails(t, rootDrops(a, roots), sys, "orderer's After of unbuilt")
	})
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
