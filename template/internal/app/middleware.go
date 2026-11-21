package app

import (
	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-web-sdk"
	mw "github.com/standards-lab/go-web-sdk/middleware"
)

func middleware(s *graph.Scope, n *Nodes) []web.Middleware {
	logger := s.Use(n.Logger)
	return []web.Middleware{
		mw.RequestID(mw.RequestIDOptions{}),
		mw.RequestLogger(logger),
		mw.Recoverer(logger),
	}
}
