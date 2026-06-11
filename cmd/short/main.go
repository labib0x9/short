package main

func main() {

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

// 	// Set up Gin router and endpoints
// 	router := gin.Default()
// 	router.GET("/:code", urlHandler.GetFullURL)                                                           // Redirect to original URL
// 	router.GET("/fetch/:code", urlHandler.GetUrlMetadata)                                            // Fetch original URL without redirect
// 	router.POST("/shorten", middleware.RateLimitByUserAgentMiddleware(redisCache), urlHandler.ShortenURL) // Create a new short URL

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
