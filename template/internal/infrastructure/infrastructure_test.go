package infrastructure_test

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

// checker is a fixed-answer readiness check for registrations under test.
type checker bool

func (c checker) Ready() bool { return bool(c) }

// store exercises registration under an interface type.
type store interface{ kind() string }

type memStore struct{}

func (memStore) kind() string { return "mem" }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := infrastructure.NewRegistry()
	logger := slog.New(slog.DiscardHandler)

	r.Register(logger, infrastructure.Service{Name: "logger"})

	if got := r.Get[*slog.Logger](); got != logger {
		t.Error("Get returned a different instance than Register stored")
	}
}

func TestRegistry_RegistersUnderAnInterfaceType(t *testing.T) {
	r := infrastructure.NewRegistry()

	r.Register[store](memStore{}, infrastructure.Service{Name: "store"})

	if got := r.Get[store]().kind(); got != "mem" {
		t.Errorf("Get[store]().kind() = %q, want mem", got)
	}
}

// writeLog and readLog are the multiplicity rule under test: two roles
// sharing an underlying type each carry a defined wrapper type, so both
// register and the role distinction stays in the type system.
type writeLog struct{ *slog.Logger }

type readLog struct{ *slog.Logger }

func TestRegister_WrapperTypesDistinguishRoles(t *testing.T) {
	r := infrastructure.NewRegistry()
	w := writeLog{slog.New(slog.DiscardHandler)}
	rd := readLog{slog.New(slog.DiscardHandler)}

	r.Register(w, infrastructure.Service{Name: "write-log"})
	r.Register(rd, infrastructure.Service{Name: "read-log"})

	if got := r.Get[writeLog](); got != w {
		t.Error("Get[writeLog] returned a different instance than Register stored")
	}
	if got := r.Get[readLog](); got != rd {
		t.Error("Get[readLog] returned a different instance than Register stored")
	}
}

func TestRegister_DuplicateTypePanics(t *testing.T) {
	r := infrastructure.NewRegistry()
	r.Register("first", infrastructure.Service{Name: "a"})

	defer func() {
		if recover() == nil {
			t.Error("registering the same type twice did not panic")
		}
	}()
	r.Register("second", infrastructure.Service{Name: "b"})
}

func TestGet_UnregisteredTypePanics(t *testing.T) {
	r := infrastructure.NewRegistry()

	defer func() {
		if recover() == nil {
			t.Error("Get of an unregistered type did not panic")
		}
	}()
	r.Get[*slog.Logger]()
}

func TestStart_RunsInRegistrationOrderSkippingNilHooks(t *testing.T) {
	r := infrastructure.NewRegistry()
	var order []string
	start := func(name string) func(context.Context) error {
		return func(context.Context) error {
			order = append(order, name)
			return nil
		}
	}

	r.Register(1, infrastructure.Service{Name: "first", Start: start("first")})
	r.Register("x", infrastructure.Service{Name: "hookless"})
	r.Register(true, infrastructure.Service{Name: "third", Start: start("third")})

	if err := r.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if want := []string{"first", "third"}; !slices.Equal(order, want) {
		t.Errorf("start order = %v, want %v", order, want)
	}
}

func TestStart_StopsAtTheFirstFailure(t *testing.T) {
	r := infrastructure.NewRegistry()
	boom := errors.New("boom")
	var reached bool

	r.Register(1, infrastructure.Service{
		Name:  "failing",
		Start: func(context.Context) error { return boom },
	})
	r.Register("x", infrastructure.Service{
		Name:  "after",
		Start: func(context.Context) error { reached = true; return nil },
	})

	err := r.Start(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("Start = %v, want the service's error wrapped", err)
	}
	if !strings.Contains(err.Error(), "failing") {
		t.Errorf("error %q does not name the failing service", err)
	}
	if reached {
		t.Error("Start continued past the failure")
	}
}

func TestShutdown_ReverseOrderJoiningEveryError(t *testing.T) {
	r := infrastructure.NewRegistry()
	var order []string
	errA, errB := errors.New("a failed"), errors.New("b failed")
	shutdown := func(name string, err error) func(context.Context) error {
		return func(context.Context) error {
			order = append(order, name)
			return err
		}
	}

	r.Register(1, infrastructure.Service{Name: "a", Shutdown: shutdown("a", errA)})
	r.Register("x", infrastructure.Service{Name: "b", Shutdown: shutdown("b", errB)})

	err := r.Shutdown(context.Background())
	if want := []string{"b", "a"}; !slices.Equal(order, want) {
		t.Errorf("shutdown order = %v, want %v", order, want)
	}
	if !errors.Is(err, errA) || !errors.Is(err, errB) {
		t.Errorf("Shutdown = %v, want both services' errors joined", err)
	}
}

func TestChecks_CollectsTheServicesThatReportReadiness(t *testing.T) {
	r := infrastructure.NewRegistry()
	r.Register(1, infrastructure.Service{Name: "silent"})
	r.Register("x", infrastructure.Service{Name: "ready", Check: checker(true)})

	checks := r.Checks()
	if len(checks) != 1 {
		t.Fatalf("Checks() = %d entries, want 1", len(checks))
	}
	if checks[0].Name != "ready" {
		t.Errorf("check name = %q, want ready", checks[0].Name)
	}
	if !checks[0].Checker.Ready() {
		t.Error("check does not report the registered readiness")
	}
}

// The baseline registers only the inert logger: the lifecycle halves must be
// clean no-ops while nothing declares hooks.
func TestEmptyRegistryLifecycleIsClean(t *testing.T) {
	r := infrastructure.NewRegistry()

	if err := r.Start(context.Background()); err != nil {
		t.Errorf("Start = %v, want nil", err)
	}
	if err := r.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown = %v, want nil", err)
	}
	if checks := r.Checks(); len(checks) != 0 {
		t.Errorf("Checks() = %v, want none", checks)
	}
}
