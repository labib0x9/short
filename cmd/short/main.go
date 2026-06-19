package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labib0x9/short/config"
	urlapp "github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/cron"
	"github.com/labib0x9/short/internal/infra/postgres"
	"github.com/labib0x9/short/internal/infra/rabbitmq"
	"github.com/labib0x9/short/internal/infra/redis"
	redis_cache "github.com/labib0x9/short/internal/infra/redis/cache"
	ratelimitter "github.com/labib0x9/short/internal/infra/redis/rate_limitter"
	rest "github.com/labib0x9/short/internal/transport/http"
	"github.com/labib0x9/short/internal/transport/http/handler/static"
	"github.com/labib0x9/short/internal/transport/http/handler/url"
	"github.com/labib0x9/short/internal/worker"
)

func main() {

	cnf := config.GetConfig()

	pgConn := postgres.NewPostgresConn(cnf.PostgreSQL)
	defer pgConn.Close()

	redisClient := redis.Setup(cnf.Redis)
	defer redisClient.Close()

	rabbitMq := rabbitmq.NewRabbitMQ(cnf.RabbitMq)
	defer rabbitMq.Close()

	urlRepo := postgres.NewUrlRepository(pgConn)
	analysisRepo := postgres.NewAnalysisRepository(pgConn)
	cacheRepo := redis_cache.NewCache(redisClient)
	rateLimiter := ratelimitter.NewRateLimiter(redisClient)

	txMngr := postgres.NewTxManager(pgConn)

	urlService := urlapp.NewService(urlRepo, analysisRepo, txMngr, cacheRepo, rabbitMq, cnf)

	worker := worker.NewWorker(rabbitMq, urlService)
	cleaner := cron.NewCleaner(urlService)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go worker.Run(ctx, "analytics-worker", 10)
	go cleaner.Run(ctx)

	validate := validator.New()
	staticHandler := static.NewHandler()
	urlHandler := url.NewHandler(urlService, validate)

	server := rest.NewServer(urlHandler, staticHandler)

	go func() {
		server.Start(rateLimiter, cnf)
	}()

	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.Shutdown(shutdown)
}
