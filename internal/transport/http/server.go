package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/domain/cache"
	"github.com/labib0x9/short/internal/transport/http/handler/static"
	"github.com/labib0x9/short/internal/transport/http/handler/url"
	"github.com/labib0x9/short/internal/transport/http/middleware"
)

type Server struct {
	urlHandler    *url.Handler
	staticHandler *static.Handler
	server        http.Server
}

func NewServer(urlHandler *url.Handler, staticHandler *static.Handler) Server {
	return Server{
		urlHandler:    urlHandler,
		staticHandler: staticHandler,
	}
}

func (s *Server) Start(limiter cache.RateLimiter, cnf *config.Config) {

	rateLimiter := middleware.NewRateLimiter(limiter, 5, 10)

	manager := middleware.NewManager()
	manager.Use(
		middleware.Logger,
		rateLimiter.Limit(),
	)

	mux := http.NewServeMux()
	wrappedMux := manager.WrapMux(mux)

	s.urlHandler.RegisterRoutes(mux, manager)
	s.staticHandler.RegisterRoutes(mux, manager)

	addr := fmt.Sprintf("http://%s:%d", cnf.Addr, cnf.Port)
	s.server = http.Server{
		Addr:    fmt.Sprintf(":%d", cnf.Port),
		Handler: wrappedMux,
	}

	fmt.Printf("Starting Server at %s\n", addr)
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server ListenAndServe():", "error", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	s.server.Shutdown(ctx)
}
