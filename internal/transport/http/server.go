package http

import (
	"context"
	"fmt"
	"log"
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

	fmt.Printf("Starting Server at http://127.0.0.1:%d/\n", cnf.Port)
	log.Fatal(http.ListenAndServe(":3000", wrappedMux))
}

func (s *Server) Shutdown(ctx context.Context) {

}
