package middleware

import "github.com/labib0x9/short/internal/infra/cache"

type Middlewares struct {
	cache cache.CacheRepo
}

func NewMiddlewares(cache cache.CacheRepo) *Middlewares {
	return &Middlewares{
		cache: cache,
	}
}
