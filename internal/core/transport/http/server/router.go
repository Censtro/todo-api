package core_http_server

import (
	"fmt"
	"net/http"

	core_http_middleware "github.com/Censtro/todo-api/internal/core/transport/http/middleware"
)

type APIVersion string

var (
	APIVersion1 = APIVersion("v1")
	APIVersion2 = APIVersion("v2")
)

type APIVersionRouter struct {
	*http.ServeMux
	apiVersion APIVersion
	middelware []core_http_middleware.Middleware
}

func NewAPIVersionRouter(
	apiVersion APIVersion,
	middelware ...core_http_middleware.Middleware,
) *APIVersionRouter {
	return &APIVersionRouter{
		ServeMux:   http.NewServeMux(),
		apiVersion: apiVersion,
		middelware: middelware,
	}
}

func (r *APIVersionRouter) RegisterRoute(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)

		r.Handle(pattern, route.WithMiddelware())
	}

}

func (r *APIVersionRouter) WithMiddelware() http.Handler {
	return core_http_middleware.ChainMiddleware(
		r,
		r.middelware...,
	)
}
