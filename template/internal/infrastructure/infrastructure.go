package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"
)

// Service is one infrastructure service's lifecycle declaration: the hooks
// the registry drives on the service's behalf. Every hook is optional — a
// nil Start or Shutdown is a no-op, and a nil Check reports no readiness.
// Name labels startup errors, shutdown errors, and the readiness probe.
type Service struct {
	Name     string
	Start    func(ctx context.Context) error
	Shutdown func(ctx context.Context) error
	Check    lifecycle.ReadinessChecker
}

// Registry holds the infrastructure services an application composes on,
// keyed by type. [Registry.Register] stores one instance per type together
// with its lifecycle declaration; [Registry.Get] retrieves it. Registration
// order is startup order, and shutdown runs in reverse. Only the composition
// root and the application layer touch the registry; everything below
// receives its dependencies through constructor parameters.
type Registry struct {
	values   map[reflect.Type]any
	services []Service
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		values: make(map[reflect.Type]any),
	}
}

// Register stores value as the registry's instance of type T and appends its
// lifecycle declaration to the startup order. Registering a type twice
// panics: one instance per type is the contract, and a duplicate is a wiring
// mistake surfaced at cold start. Two roles sharing a type (a write pool and
// a read pool) register as defined wrapper types, so the role distinction
// stays in the type system.
func (r *Registry) Register[T any](value T, svc Service) {
	key := reflect.TypeFor[T]()
	if _, exists := r.values[key]; exists {
		panic(fmt.Sprintf("infrastructure: %s already registered", key))
	}
	r.values[key] = value
	r.services = append(r.services, svc)
}

// Get returns the registry's instance of type T, panicking when none is
// registered — a wiring mistake, surfaced at cold start before any I/O.
func (r *Registry) Get[T any]() T {
	value, ok := r.values[reflect.TypeFor[T]()]
	if !ok {
		panic(fmt.Sprintf("infrastructure: %s not registered", reflect.TypeFor[T]()))
	}
	return value.(T)
}

// Start establishes connectivity in registration order, stopping at the
// first failure. It carries the lifecycle hook signature.
func (r *Registry) Start(ctx context.Context) error {
	for _, svc := range r.services {
		if svc.Start == nil {
			continue
		}
		if err := svc.Start(ctx); err != nil {
			return fmt.Errorf("%s: %w", svc.Name, err)
		}
	}
	return nil
}

// Shutdown tears the services down in reverse registration order, so a
// dependent drains before what it depends on. Every service gets its
// shutdown attempt; the errors join.
func (r *Registry) Shutdown(ctx context.Context) error {
	var errs []error
	for _, svc := range slices.Backward(r.services) {
		if svc.Shutdown == nil {
			continue
		}
		if err := svc.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", svc.Name, err))
		}
	}
	return errors.Join(errs...)
}

// Checks yields the readiness checks of the services that report readiness,
// in registration order, named for the probe.
func (r *Registry) Checks() []web.Check {
	var checks []web.Check
	for _, svc := range r.services {
		if svc.Check == nil {
			continue
		}
		checks = append(checks, web.Check{Name: svc.Name, Checker: svc.Check})
	}
	return checks
}
