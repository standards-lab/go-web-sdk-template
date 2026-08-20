package main

import (
	"io"

	"github.com/standards-lab/go-core/logging"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
)

func setInfrastructure(
	w io.Writer,
	cfg *config.Config,
) (*infrastructure.Registry, error) {
	r := infrastructure.NewRegistry()

	r.Register(logging.New(w, cfg.Log), infrastructure.Service{
		Name: "logger",
	})

	return r, nil
}
