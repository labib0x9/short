package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labib0x9/short/config"
	urlapp "github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/infra/postgres"
	"github.com/labib0x9/short/internal/infra/rabbitmq"
	"github.com/labib0x9/short/internal/infra/redis"
	rest "github.com/labib0x9/short/internal/transport/http"
	"github.com/labib0x9/short/internal/transport/http/handler/static"
	"github.com/labib0x9/short/internal/transport/http/handler/url"
	"github.com/labib0x9/short/internal/worker"
)

func main() {

	cnf := config.GetConfig()

	postgresConn := postgres.New()
	dbConn := postgresConn.SetupAndConnection(cnf.PostgreSQL)
	defer dbConn.Close()

	redisClient := redis.Setup(cnf.Redis)
	defer redisClient.Close()

	rabbitMq := rabbitmq.NewRabbitMQ(cnf.RabbitMq)
	defer rabbitMq.Close()

	urlRepo := postgres.NewUrlRepository(dbConn)
	analysisRepo := postgres.NewAnalysisRepository(dbConn)
	cacheRepo := redis.NewCacheRepo(redisClient)

	// middlewares := middleware.NewMiddlewares(cacheRepo)

	worker := worker.NewWorker(rabbitMq, urlRepo, analysisRepo, cacheRepo)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go worker.Run(ctx, 10)

	urlService := urlapp.NewService(urlRepo, analysisRepo, cacheRepo, cnf)

	validate := validator.New()
	staticHandler := static.NewHandler()
	urlHandler := url.NewHandler(urlService, rabbitMq, validate)

	server := rest.NewServer(urlHandler, staticHandler)

	go func() {
		server.Start(redisClient, cnf)
	}()

	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.Shutdown(shutdown)
}
