package url

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/labib0x9/short/internal/domain/queue"
	"github.com/labib0x9/short/internal/domain/url"
)

func (s *service) Get(ctx context.Context, code, referer, userAgent, remoteAddr string) (*url.Url, error) {
	now := time.Now()
	expireKey := "expire:" + code
	_, err := s.cache.Get(ctx, expireKey)
	if err == nil {
		return nil, url.ErrShortCodeExpired
	}

	cacheKey := "short:" + code
	value, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var cached url.Url
		if err := json.Unmarshal([]byte(value), &cached); err == nil {
			go s.publish(code, now, referer, userAgent, remoteAddr)
			return &cached, nil
		}
	}

	fetchedUrl, err := s.urlRepo.GetByShortCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, url.ErrUrlNotFound
		}
		return nil, err
	}

	if fetchedUrl.IsExpired() {
		s.cache.Set(ctx, expireKey, "1", 0)
		return nil, url.ErrShortCodeExpired
	}

	duration := 5 * time.Minute
	if fetchedUrl.ExpireAt != nil {
		expire := time.Until(*fetchedUrl.ExpireAt)
		duration = min(expire, duration)
	}

	urlJson, err := json.Marshal(fetchedUrl)
	s.cache.Set(ctx, cacheKey, string(urlJson), duration)

	go s.publish(code, now, referer, userAgent, remoteAddr)

	return fetchedUrl, err
}

func (s *service) publish(code string, createdAt time.Time, referer, userAgent, remoteAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.queue.PublishAnalytics(ctx, queue.ClickEvent{
		ShortCode: code,
		ClickedAt: createdAt,
		Referer:   referer,
		UserAgent: userAgent,
		IP:        remoteAddr,
		Retries:   2,
	}); err != nil {
		slog.Error("Publishing failed", "error", err, "short", code)
	}
}
