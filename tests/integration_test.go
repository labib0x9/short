//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/app/url"
	"github.com/labib0x9/short/internal/infra/postgres"
	"github.com/labib0x9/short/internal/infra/redis"
	redis_cache "github.com/labib0x9/short/internal/infra/redis/cache"
)

func TestShorten_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup
	cnf := &config.Config{
		PostgreSQL: &config.PostgreSQL{
			User:         "short",
			Pass:         "secret",
			Port:         "5432",
			Addr:         "localhost",
			DatabaseName: "short",
			SslMode:      "disable",
		},
		Redis: &config.Redis{Addr: "localhost:6379"},
		Prefix: "http://localhost:3000/",
	}

	pgConn := postgres.NewPostgresConn(cnf.PostgreSQL)
	defer pgConn.Close()

	redisClient := redis.Setup(cnf.Redis)
	defer redisClient.Close()

	urlRepo := postgres.NewUrlRepository(pgConn)
	analysisRepo := postgres.NewAnalysisRepository(pgConn)
	cacheRepo := redis_cache.NewCache(redisClient)
	txMngr := postgres.NewTxManager(pgConn)

	svc := url.NewService(urlRepo, analysisRepo, txMngr, cacheRepo, nil, cnf)

	// Test: Shorten a URL
	ctx := context.Background()
	result, err := svc.Shorten(ctx, "https://example.com/very/long/path", nil, "")
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if result.Code == "" {
		t.Error("expected non-empty code")
	}

	if result.Msg != "success" {
		t.Errorf("expected msg=success, got %s", result.Msg)
	}

	if result.ShortUrl == "" {
		t.Error("expected non-empty short_url")
	}
}

func TestFetch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup (reuse from above)
	cnf := &config.Config{
		PostgreSQL: &config.PostgreSQL{
			User:         "short",
			Pass:         "secret",
			Port:         "5432",
			Addr:         "localhost",
			DatabaseName: "short",
			SslMode:      "disable",
		},
		Redis: &config.Redis{Addr: "localhost:6379"},
		Prefix: "http://localhost:3000/",
	}

	pgConn := postgres.NewPostgresConn(cnf.PostgreSQL)
	defer pgConn.Close()

	redisClient := redis.Setup(cnf.Redis)
	defer redisClient.Close()

	urlRepo := postgres.NewUrlRepository(pgConn)
	analysisRepo := postgres.NewAnalysisRepository(pgConn)
	cacheRepo := redis_cache.NewCache(redisClient)
	txMngr := postgres.NewTxManager(pgConn)

	svc := url.NewService(urlRepo, analysisRepo, txMngr, cacheRepo, nil, cnf)

	// Test: Create and fetch URL
	ctx := context.Background()

	// Create
	result, err := svc.Shorten(ctx, "https://example.com/path", nil, "")
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Fetch
	fetchedUrl, err := svc.Get(ctx, result.Code, "https://referrer.com", "Mozilla/5.0", "127.0.0.1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if fetchedUrl.URL != "https://example.com/path" {
		t.Errorf("expected url=https://example.com/path, got %s", fetchedUrl.URL)
	}
}

func TestAnalysis_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup
	cnf := &config.Config{
		PostgreSQL: &config.PostgreSQL{
			User:         "short",
			Pass:         "secret",
			Port:         "5432",
			Addr:         "localhost",
			DatabaseName: "short",
			SslMode:      "disable",
		},
		Redis: &config.Redis{Addr: "localhost:6379"},
		Prefix: "http://localhost:3000/",
	}

	pgConn := postgres.NewPostgresConn(cnf.PostgreSQL)
	defer pgConn.Close()

	redisClient := redis.Setup(cnf.Redis)
	defer redisClient.Close()

	urlRepo := postgres.NewUrlRepository(pgConn)
	analysisRepo := postgres.NewAnalysisRepository(pgConn)
	cacheRepo := redis_cache.NewCache(redisClient)
	txMngr := postgres.NewTxManager(pgConn)

	svc := url.NewService(urlRepo, analysisRepo, txMngr, cacheRepo, nil, cnf)

	ctx := context.Background()

	// Create URL
	result, err := svc.Shorten(ctx, "https://example.com/analytics-test", nil, "")
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Get analytics (should have 0 clicks initially)
	analysis, err := svc.Analysis(ctx, result.Code)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	if analysis.ClickCount != 0 {
		t.Errorf("expected initial click count=0, got %d", analysis.ClickCount)
	}

	if analysis.ShortURL != result.Code {
		t.Errorf("expected short=%s, got %s", result.Code, analysis.ShortURL)
	}
}

func TestURLExpiration_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup
	cnf := &config.Config{
		PostgreSQL: &config.PostgreSQL{
			User:         "short",
			Pass:         "secret",
			Port:         "5432",
			Addr:         "localhost",
			DatabaseName: "short",
			SslMode:      "disable",
		},
		Redis: &config.Redis{Addr: "localhost:6379"},
		Prefix: "http://localhost:3000/",
	}

	pgConn := postgres.NewPostgresConn(cnf.PostgreSQL)
	defer pgConn.Close()

	redisClient := redis.Setup(cnf.Redis)
	defer redisClient.Close()

	urlRepo := postgres.NewUrlRepository(pgConn)
	analysisRepo := postgres.NewAnalysisRepository(pgConn)
	cacheRepo := redis_cache.NewCache(redisClient)
	txMngr := postgres.NewTxManager(pgConn)

	svc := url.NewService(urlRepo, analysisRepo, txMngr, cacheRepo, nil, cnf)

	ctx := context.Background()

	// Create URL with expiration in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	_, err := svc.Shorten(ctx, "https://example.com", &pastTime, "")

	// Should fail because expire_at is in the past
	if err == nil {
		t.Error("expected error for past expire_at")
	}
}
