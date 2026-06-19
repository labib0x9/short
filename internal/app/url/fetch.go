package url

import (
	"context"
	"encoding/json"
	"time"

	"github.com/labib0x9/short/internal/domain/url"
)

func (s *service) Get(ctx context.Context, code string) (*url.Url, error) {
	expireKey := "expire:" + code
	_, err := s.cache.Get(ctx, expireKey)
	if err == nil {
		return nil, url.ErrShortCodeExpired
	}

	cacheKey := "short:" + code
	value, err := s.cache.Get(ctx, code)
	if err == nil {
		var cached url.Url
		if err := json.Unmarshal([]byte(value), &cached); err == nil {

			return &cached, nil
		}
	}

	fetchedUrl, err := s.urlRepo.GetByShortCode(ctx, code)
	if err != nil || fetchedUrl == nil {
		return nil, err
	}

	duration := 5 * time.Minute
	if fetchedUrl.ExpireAt != nil {
		if fetchedUrl.CreatedAt != *fetchedUrl.ExpireAt && fetchedUrl.ExpireAt.Before(time.Now()) {
			s.cache.Set(ctx, expireKey, "1", 0)
			return nil, url.ErrShortCodeCollision
		}
		if err == nil {
			expire := fetchedUrl.ExpireAt.Sub(fetchedUrl.CreatedAt)
			duration = min(expire, duration)
		}
	}

	urlJson, err := json.Marshal(fetchedUrl)

	s.cache.Set(ctx, cacheKey, string(urlJson), duration)
	return fetchedUrl, err
}
