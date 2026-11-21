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

func TestGraph_NodesNamedByTheirFields(t *testing.T) {
	a := app.New(configtest.Config(t), io.Discard)
	graphtest.CheckNames(t, a.Nodes())
}

func runSet(a *app.App, extra ...graph.Ref) []graph.Ref {
	return append(append(a.Roots(), a.Nodes().LifecycleConfig), extra...)
}

func buildAsRun(t *testing.T, a *app.App, extra ...graph.Ref) *graph.System {
	t.Helper()
	sys, err := a.Graph().Build(runSet(a, extra...)...)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return sys
}

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

func layerError(sys *graph.System, format string, args ...any) error {
	return fmt.Errorf(format+"\nlayers:\n%s", append(args, sys)...)
}

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

func fixtureReactor(a *app.App, name string, opts ...graph.Option) *graph.Node[*reactor.Reactor[time.Time]] {
	opts = append([]graph.Option{graph.After(a.Nodes().Router)}, opts...)
	return a.Graph().Define(name, func(*graph.Scope) (*reactor.Reactor[time.Time], error) {
		return reactor.New(reactor.Every(time.Hour), func(context.Context, time.Time) error {
			return nil
		}, reactor.Options{}), nil
	}, opts...)
}

func TestGraph_ServerAboveReactor(t *testing.T) {
	t.Run("in the reactors group", func(t *testing.T) {
		a := app.New(configtest.Config(t), io.Discard)
		fixture := fixtureReactor(a, "fixture", graph.In(a.Nodes().Reactors))

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

func TestRun_BuildFailureExitsOne(t *testing.T) {
	var log, stderr syncBuffer
	a := app.New(configtest.Config(t), &log)
	a.Graph().Replace(a.Nodes().Logger, func(*graph.Scope) (*slog.Logger, error) {
		return nil, errors.New("unavailable")
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if code := a.Run(ctx, &stderr); code != 1 {
		t.Errorf("Run = %d, want 1", code)
	}
	failure := regexp.MustCompile(`^service failed: .*\blogger\b.*unavailable`)
	if got := stderr.String(); !failure.MatchString(got) {
		t.Errorf("stderr = %q, want a service failure naming the logger node and its error", got)
	}
	if got := log.String(); got != "" {
		t.Errorf("log = %q, want nothing logged", got)
	}
}
