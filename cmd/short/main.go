package main

import (
	"github.com/labib0x9/short/config"
	// urlapp "github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/infra/postgres"
	"github.com/labib0x9/short/internal/infra/rabbitmq"
	"github.com/labib0x9/short/internal/infra/redis"
	// rest "github.com/labib0x9/short/internal/transport/http"
	// "github.com/labib0x9/short/internal/transport/http/handler/url"
	// "github.com/labib0x9/short/internal/worker"
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

	// urlRepo := postgres.NewUrlRepository(dbConn)
	// analysisRepo := postgres.NewAnalysisRepository(dbConn)
	// cacheRepo := redis.NewCacheRepo(redisClient)

	// // middlewares := middleware.NewMiddlewares(cacheRepo)

	// worker := worker.NewWorker(rabbitMq, cacheRepo)

	// ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// defer stop()

	// go worker.Run(ctx, 10)

	// urlService := urlapp.NewService(urlRepo, analysisRepo)

	// urlHandler := url.NewHandler(&urlService)

	// server := rest.NewServer(urlHandler)

	// server.Start(redisClient, cnf)

}

// func main() {
// 	// Set Gin to release mode for production
// 	gin.SetMode(gin.ReleaseMode)

// 	// Load environment variables from .env file
// 	err := godotenv.Load()
// 	if err != nil {
// 		panic(err) // Panic if .env loading fails
// 	}

// 	// Initialize MySQL database connection
// 	db, err := db.NewMysqlDb()
// 	if err != nil {
// 		panic(err) // Panic if DB connection fails
// 	}
// 	defer db.Close() // Ensure DB connection is closed on exit

// 	// Initialize Redis client
// 	redis, err := Redis.NewRedisClient(0)
// 	if err != nil {
// 		panic(err) // Panic if Redis connection fails
// 	}
// 	defer redis.Close() // Ensure Redis connection is closed on exit

// 	// Create Redis cache wrapper
// 	redisCache := cache.NewRedisCache(redis)

// 	// Set up repositories and services
// 	mysqlUrlRepo := repositories.NewMysqlUrlRepository(db)
// 	redisMysqlUrlRepo := repositories.NewRedisMysqlUrlRepository(mysqlUrlRepo, redisCache)
// 	urlService := services.NewUrlService(redisMysqlUrlRepo)
// 	urlHandler := handlers.NewShortenHandler(urlService)

// 	// Build server address from environment variables
// 	host := os.Getenv("SERVER_HOST")
// 	port := os.Getenv("PORT")
// 	addr := net.JoinHostPort(host, port)

// 	// Create HTTP server
// 	server := &http.Server{
// 		Addr:    addr,
// 		Handler: router,
// 	}

// 	// Channel to receive server errors
// 	serverErrorCh := make(chan error, 1)
// 	go func() {
// 		serverErrorCh <- server.ListenAndServe() // Start HTTP server
// 	}()

// 	// Channel to listen for OS interrupt or terminate signals
// 	exitCh := make(chan os.Signal, 1)
// 	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

// 	// Wait for either a server error or a shutdown signal
// 	select {
// 	case <-serverErrorCh:
// 		// Server error occurred
// 	case <-exitCh:
// 		// Received shutdown signal
// 	}

// 	// Gracefully shutdown the server with a timeout
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	server.Shutdown(ctx)
// }
