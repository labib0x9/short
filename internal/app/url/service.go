package url

import (
	"context"
	"time"

	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/domain/cache"
	"github.com/labib0x9/short/internal/domain/db"
	"github.com/labib0x9/short/internal/domain/queue"
	"github.com/labib0x9/short/internal/domain/url"
)

type Service interface {
	Shorten(ctx context.Context, longUrl string, expireAt *time.Time, userAgent string) (ShortenResult, error)
	Get(ctx context.Context, code, referer, userAgent, remoteAddr string) (*url.Url, error)
	Analysis(ctx context.Context, code string) (url.Analysis, error)
	Save(ctx context.Context, msg queue.ClickEvent) error
	DeleteByExpireAt(ctx context.Context) error
}

type service struct {
	urlRepo      url.UrlRepository
	analysisRepo url.AnalyticsRepository
	txMngr       db.TxManager // transaction manager
	cache        cache.Cache
	queue        queue.Queue
	cnf          *config.Config
}

func NewService(
	urlRepo url.UrlRepository,
	analysisRepo url.AnalyticsRepository,
	txMngr db.TxManager,
	cache cache.Cache,
	queue queue.Queue,
	cnf *config.Config,
) Service {
	return &service{
		urlRepo:      urlRepo,
		analysisRepo: analysisRepo,
		txMngr:       txMngr,
		cache:        cache,
		queue:        queue,
		cnf:          cnf,
	}
}
