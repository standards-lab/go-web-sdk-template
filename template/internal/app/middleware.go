package app

import (
	"github.com/standards-lab/go-web-sdk"
	mw "github.com/standards-lab/go-web-sdk/middleware"
)

// middleware declares the router-level stack, outermost first. It takes
// infra, not dom: request logging, and cross-cutting concerns like it, need
// infrastructure primitives, not domain services. A middleware that has to
// reach a domain service is domain logic, and belongs in a route or a
// reactor instead.
//
// RequestID runs ahead of RequestLogger, so the log record carries the
// request id. Recoverer turns a handler's panic into a logged 500, which
// RequestLogger records with its status; it works on either side of
// RequestLogger, and the template places it inside. A
// BodyLimit, when a service adds one, goes ahead of RequestLogger and
// Recoverer; the chain order in go-web-sdk's middleware package
// documentation states why.
func middleware(infra *Infrastructure) []web.Middleware {
	return []web.Middleware{
		mw.RequestID(),
		mw.RequestLogger(infra.Logger),
		mw.Recoverer(infra.Logger),
	}
}
