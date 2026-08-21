package infrastructure

import (
	"io"
	"log/slog"

	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-core/logging"
	"github.com/standards-lab/go-web-sdk-template/template/internal/config"
)

type Infrastructure struct {
	Logger *slog.Logger
}

func New(
	w io.Writer,
	cfg *config.Config,
	lc *lifecycle.Coordinator,
) (*Infrastructure, error) {
	logger := logging.New(w, cfg.Log)

	return &Infrastructure{
		Logger: logger,
	}, nil
}
