package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	core_logger "github.com/Censtro/todo-api/internal/core/logger"
	core_http_middleware "github.com/Censtro/todo-api/internal/core/transport/http/middleware"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux *http.ServeMux
	cfg Config
	log *core_logger.Logger

	middleware []core_http_middleware.Middleware
}

func NewHTTPServer(
	cfg Config,
	log *core_logger.Logger,
	middleware ...core_http_middleware.Middleware,
) *HTTPServer {
	return &HTTPServer{
		mux:        http.NewServeMux(),
		cfg:        cfg,
		log:        log,
		middleware: middleware,
	}
}

func (h *HTTPServer) RegisterAPIRouters(routers ...*APIVersionRouter) {
	for _, router := range routers {
		prefix := "/api/" + string(router.apiVersion)
		h.mux.Handle(
			prefix+"/",
			http.StripPrefix(prefix, router),
		)
	}
}

func (h *HTTPServer) Run(ctx context.Context) error {
	mux := core_http_middleware.ChainMiddleware(h.mux, h.middleware...)

	server := &http.Server{
		Addr:    h.cfg.Address,
		Handler: mux,
	}
	ch := make(chan error, 1)
	go func() {
		defer close(ch)

		h.log.Warn("start HTTP server", zap.String("addr", h.cfg.Address))

		err := server.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()
	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("listen and serve HTTP error: %w", err)
		}
	case <-ctx.Done():
		h.log.Warn("HTTP server shutdown...")

		shutdownContext, cancel := context.WithTimeout(
			context.Background(),
			h.cfg.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			_ = server.Close()

			return fmt.Errorf("Shutdown HTTP server: %w", err)
		}

		h.log.Warn("Server Stopped")
	}

	return nil
}
