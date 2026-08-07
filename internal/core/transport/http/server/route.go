package core_http_server

import (
	"net/http"

	core_http_middleware "github.com/Censtro/todo-api/internal/core/transport/http/middleware"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middelware []core_http_middleware.Middleware
}

func (r *Route) WithMiddelware() http.Handler {
	return core_http_middleware.ChainMiddleware(
		r.Handler,
		r.Middelware...,
	)
}

func NewRoute(
	method string,
	path string,
	handler http.HandlerFunc,
	middelware ...core_http_middleware.Middleware,
) Route {
	return Route{
		Method:     method,
		Path:       path,
		Handler:    handler,
		Middelware: middelware,
	}
}
