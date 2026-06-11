package redis

import (
	"context"
	"time"

	"github.com/labib0x9/short/internal/infra/cache"
)

type cacheRepo struct {
	cache *Redis
}

func NewCacheRepo(
	cache *Redis,
) cache.CacheRepo {
	return &cacheRepo{
		cache: cache,
	}
}

func (r *cacheRepo) Set(ctx context.Context, key string, value string, expire time.Duration) error {
	return r.cache.Client.Set(
		ctx,
		key,
		value,
		expire,
	).Err()
}

func (r *cacheRepo) Get(ctx context.Context, key string) (string, error) {
	return r.cache.Client.Get(ctx, key).Result()
}

// package cache

// import (
// 	"context"
// 	"time"

// 	"github.com/redis/go-redis/v9"
// )

// // RedisCache implements the Cache interface using a Redis backend.
// // It provides methods to get and set key-value pairs in Redis with optional expiration.
// type RedisCache struct {
// 	redis *redis.Client   // Redis client instance
// 	ctx   context.Context // Context for Redis operations
// }

// // NewRedisCache creates a new RedisCache with the given Redis client.
// func NewRedisCache(redis *redis.Client) *RedisCache {
// 	return &RedisCache{
// 		redis: redis,
// 		ctx:   context.Background(),
// 	}
// }

// // Get retrieves the value for a given key from Redis.
// // Returns the value as a string or an error if the key does not exist or on failure.
// func (r *RedisCache) Get(key string) (string, error) {
// 	return r.redis.Get(r.ctx, key).Result()
// }

// // Set stores a key-value pair in Redis with the specified expiration duration.
// // If expire is 0, the key does not expire.
// func (r *RedisCache) Set(key string, value string, expire time.Duration) {
// 	r.redis.Set(r.ctx, key, value, expire)
// }

// // Incr atomically increments the integer value of a key by one in Redis.
// // Returns the new value as int64 or an error if the operation fails.
// func (r *RedisCache) Incr(key string) (int64, error) {
// 	return r.redis.Incr(r.ctx, key).Result()
// }

// // Expire sets a timeout on a key in Redis. After the timeout, the key will be automatically deleted.
// // If expire is 0, the key will not expire.
// func (r *RedisCache) Expire(key string, expire time.Duration) {
// 	r.redis.Expire(r.ctx, key, expire).Result()
// }

// package repositories

// import (
// 	"encoding/json"
// 	"log/slog"
// 	"time"
// 	"urlshortener/cache"
// 	"urlshortener/models"
// 	"urlshortener/utils"
// )

// // RedisMysqlUrlRepository is a URL repository that uses both a persistent backend (MySQL) and a cache (Redis)
// type RedisMysqlUrlRepository struct {
// 	repo  UrlRepository // Underlying persistent repository (e.g., MySQL)
// 	redis cache.Cache   // Cache layer (e.g., Redis)
// }

// // NewRedisMysqlUrlRepository creates a new RedisMysqlUrlRepository with the given persistent repo and cache
// func NewRedisMysqlUrlRepository(repo UrlRepository, redis cache.Cache) *RedisMysqlUrlRepository {
// 	return &RedisMysqlUrlRepository{
// 		repo:  repo,
// 		redis: redis,
// 	}
// }

// // Create stores a new URL mapping in the persistent repository
// func (r *RedisMysqlUrlRepository) Create(url models.Url) error {
// 	return r.repo.Create(url)
// }

// // GetByShortCode retrieves a URL by its short code, using cache and expiration logic
// func (r *RedisMysqlUrlRepository) GetByShortCode(shortCode string) (*models.Url, error) {
// 	// Check if the short code is marked as expired in cache
// 	expireKey := "expire:" + shortCode
// 	_, err := r.redis.Get(expireKey)
// 	if err == nil {
// 		slog.Error(" [redis_mysql_url_repository.go] [Expired short code] ", slog.Any("error", err))
// 		return nil, utils.ErrShortCodeExpired
// 	}

// 	// Try to get the URL from cache
// 	cacheKey := "short:" + shortCode
// 	value, err := r.redis.Get(cacheKey)
// 	if err == nil {
// 		var cacheUrl models.Url
// 		if err := json.Unmarshal([]byte(value), &cacheUrl); err == nil {
// 			return &cacheUrl, nil // Return cached URL if found
// 		}
// 	}

// 	// slog.Info(" [Cache miss] ", slog.String("shortCode", shortCode))

// 	// Fallback to persistent repository if not in cache
// 	url, err := r.repo.GetByShortCode(shortCode)
// 	if err != nil || url == nil {
// 		return nil, err
// 	}

// 	// slog.Info(" [URL found in repo] ", slog.String("shortCode", shortCode))

// 	// If the URL is expired, mark it as expired in cache and return error
// 	if url.CreatedAt != url.Expire && url.Expire.Before(time.Now()) {
// 		expireKey := "expire:" + shortCode
// 		r.redis.Set(expireKey, "1", 0) // 0 means never expire in cache
// 		return nil, utils.ErrShortCodeExpired
// 	}

// 	// Cache the URL for future lookups min(5 minutes, actual expiration)
// 	urlJson, err := json.Marshal(url)
// 	if err == nil {

// 		// slog.Info(" [Created At]", url.CreatedAt.GoString())
// 		// slog.Info(" [Expire At]", url.Expire.GoString())

// 		duration := 5 * time.Minute
// 		if url.CreatedAt != url.Expire {
// 			expire := url.Expire.Sub(url.CreatedAt)
// 			duration = min(expire, duration)
// 		}

// 		// slog.Info(" [Caching URL] ", slog.String("shortCode", shortCode), slog.String("duration", duration.String()))

// 		r.redis.Set(cacheKey, string(urlJson), duration)
// 	}

// 	return url, err
// }
