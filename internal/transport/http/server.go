package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/infra/redis"
	"github.com/labib0x9/short/internal/transport/http/handler/url"
	"github.com/labib0x9/short/internal/transport/http/middleware"
)

type Server struct {
	urlHandler *url.Handler
}

func NewServer(urlHandler *url.Handler) Server {
	return Server{
		urlHandler: urlHandler,
	}
}

func (s *Server) Start(redisClient *redis.Redis, cnf *config.Config) {

	rateLimiter := middleware.NewRateLimiter(redisClient, 5, 10)

	manager := middleware.NewManager()
	manager.Use(
		middleware.Logger,
		rateLimiter.Limit(),
	)

	mux := http.NewServeMux()
	wrappedMux := manager.WrapMux(mux)

	s.urlHandler.RegisterRoutes(mux, manager)

	fmt.Printf("Starting Server at http://127.0.0.1:%d/\n", cnf.Port)
	log.Fatal(http.ListenAndServe(":8080", wrappedMux))
}
