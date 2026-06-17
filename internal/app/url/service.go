package url

import (
	"context"
	"time"

	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/domain/cache"
	"github.com/labib0x9/short/internal/domain/queue"
	"github.com/labib0x9/short/internal/domain/url"
)

type Service interface {
	Shorten(longUrl string, expireAt *time.Time, userAgent string) (ShortenResult, error)
	Get(ctx context.Context, code string) (*url.Url, error)
	Analysis(code string) (url.Analysis, error)
	Save(ctx context.Context, msg queue.ClickEvent) error
}

type service struct {
	urlRepo      url.UrlRepository
	analysisRepo url.AnalyticsRepository
	cache        cache.Cache
	cnf          *config.Config
}

func NewService(
	urlRepo url.UrlRepository,
	analysisRepo url.AnalyticsRepository,
	cache cache.Cache,
	cnf *config.Config,
) Service {
	return &service{
		urlRepo:      urlRepo,
		analysisRepo: analysisRepo,
		cache:        cache,
		cnf:          cnf,
	}
}
