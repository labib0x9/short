package middleware

import "github.com/labib0x9/short/internal/domain/cache"

type Middlewares struct {
	cache cache.Cache
}

func NewMiddlewares(cache cache.Cache) *Middlewares {
	return &Middlewares{
		cache: cache,
	}
}
