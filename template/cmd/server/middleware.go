package main

import (
	"log/slog"

	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
	"github.com/standards-lab/go-web-sdk/middleware"
)

func setMiddleware(infra *infrastructure.Registry) []web.Middleware {
	return []web.Middleware{
		middleware.RequestLogger(infra.Get[*slog.Logger]()),
	}
}
