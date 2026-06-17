package ratelimitter

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/infra/redis"
)

// TestRateLimiterConcurrent fires N goroutines at the same key at the same
// instant and verifies the script never admits more requests than capacity
// allows, even under concurrent access. This checks the atomicity of the
// Lua script (EVAL), not your HTTP layer.
func TestRateLimiterConcurrent(t *testing.T) {
	client := redis.Setup(&config.Redis{Addr: "localhost:6379"})
	defer client.Close()

	rl := NewRateLimiter(client)
	ctx := context.Background()

	key := "test:rl:concurrent:" + uuid.NewString()
	capacity := 5
	rate := 1 // slow refill so refill doesn't muddy the result during the burst
	now := time.Now().UnixMilli()

	const goroutines = 50

	var (
		wg         sync.WaitGroup
		allowedCnt int64
		deniedCnt  int64
		errCnt     int64
	)

	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // release all goroutines at once

			res, err := rl.RunScript(ctx, key, capacity, rate, now)
			if err != nil {
				atomic.AddInt64(&errCnt, 1)
				return
			}

			data, ok := res.([]interface{})
			if !ok || len(data) < 1 {
				atomic.AddInt64(&errCnt, 1)
				return
			}

			allowed, ok := data[0].(int64)
			if !ok {
				atomic.AddInt64(&errCnt, 1)
				return
			}

			if allowed == 1 {
				atomic.AddInt64(&allowedCnt, 1)
			} else {
				atomic.AddInt64(&deniedCnt, 1)
			}
		}()
	}

	close(start) // fire all goroutines simultaneously
	wg.Wait()

	t.Logf("goroutines=%d allowed=%d denied=%d errors=%d (capacity=%d)",
		goroutines, allowedCnt, deniedCnt, errCnt, capacity)

	if errCnt > 0 {
		t.Fatalf("expected zero errors during concurrent run, got %d", errCnt)
	}

	// The critical assertion: under no circumstance should more requests be
	// admitted than the bucket capacity, regardless of concurrency.
	if allowedCnt > int64(capacity) {
		t.Fatalf("race condition detected: %d requests allowed, expected at most %d (capacity)", allowedCnt, capacity)
	}

	if allowedCnt+deniedCnt != int64(goroutines) {
		t.Fatalf("lost requests: allowed(%d)+denied(%d) != goroutines(%d)", allowedCnt, deniedCnt, goroutines)
	}
}

// TestRateLimiterConcurrentDifferentKeys verifies that concurrent requests
// to DIFFERENT keys don't interfere with each other (each bucket is
// independent), and all should be allowed since each has full capacity.
func TestRateLimiterConcurrentDifferentKeys(t *testing.T) {
	client := redis.Setup(&config.Redis{Addr: "localhost:6379"})
	defer client.Close()

	rl := NewRateLimiter(client)
	ctx := context.Background()

	capacity := 5
	rate := 10
	now := time.Now().UnixMilli()

	const goroutines = 30
	var wg sync.WaitGroup
	var allowedCnt int64

	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start

			key := "test:rl:isolated:" + uuid.NewString() // unique key per goroutine
			res, err := rl.RunScript(ctx, key, capacity, rate, now)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			data := res.([]interface{})
			if data[0].(int64) == 1 {
				atomic.AddInt64(&allowedCnt, 1)
			}
		}(i)
	}

	close(start)
	wg.Wait()

	if allowedCnt != int64(goroutines) {
		t.Fatalf("expected all %d requests on distinct keys to be allowed, got %d", goroutines, allowedCnt)
	}
}
