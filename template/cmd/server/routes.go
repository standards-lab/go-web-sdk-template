package main

import (
	"net/http"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
	"github.com/standards-lab/go-web-sdk/middleware"
)

func newRouter(
	infra *infrastructure.Infrastructure,
	lc *lifecycle.Coordinator,
) http.Handler {
	router := web.NewRouter()
	router.Use(middleware.RequestLogger(infra.Logger))

	checks := append(
		[]web.Check{{Name: "lifecycle", Checker: lc}},
		infra.Checks()...,
	)
	web.RegisterHealth(router, checks...)

	return router
}
