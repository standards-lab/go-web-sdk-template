package app

import (
	"github.com/standards-lab/go-web-sdk"
	"github.com/standards-lab/go-web-sdk-template/template/internal/infrastructure"
	mw "github.com/standards-lab/go-web-sdk/middleware"
)

func middleware(infra *infrastructure.Infrastructure) []web.Middleware {
	return []web.Middleware{
		mw.RequestLogger(infra.Logger),
	}
}
