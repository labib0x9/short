package cache

import (
	"context"
	"testing"
	"time"

	go_redis "github.com/redis/go-redis/v9"
)

func TestCache(t *testing.T) {

	// cnf := config.GetConfig()
	client := go_redis.NewClient(
		&go_redis.Options{
			Addr: "localhost:6379",
			// Username: cnf.User,
			// Password: cnf.Pass,
		},
	)

	c := NewCache(client)
	ctx := context.Background()

	err := c.Set(ctx, "foo", "bar", 5*time.Second)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := c.Get(ctx, "foo")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "bar" {
		t.Fatalf("expected bar, got %s", val)
	}

	time.Sleep(6 * time.Second)
	_, err = c.Get(ctx, "foo")
	if err == nil {
		t.Fatal("expected key to expire, but Get succeeded")
	}
}
