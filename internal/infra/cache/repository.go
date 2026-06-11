package cache

import (
	"context"
	"time"
)

type CacheRepo interface {
	Set(ctx context.Context, key string, value string, expire time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type Cache interface {
	// // Get retrieves the value for a given key.
	// Get(key string) (string, error)
	// // Set stores a key-value pair with an optional expiration duration.
	// Set(key string, value string, expire time.Duration)
	// // Incr atomically increments the integer value of a key by one.
	// Incr(key string) (int64, error)
	// // Expire sets a timeout on a key. After the timeout, the key will be automatically deleted.
	// Expire(key string, expire time.Duration)
}
