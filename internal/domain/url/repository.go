package url

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UrlRepository interface {
	Create(ctx context.Context, url Url) error
	GetByShortCode(ctx context.Context, shortCode string) (*Url, error)
	Update(ctx context.Context, id uuid.UUID, lastClickedAt time.Time) error
	GetMetadata(ctx context.Context, code string) (*Url, error)
}

type AnalyticsRepository interface {
	Create(ctx context.Context, click Click) error
	GetBrowserCount(ctx context.Context, Id uuid.UUID) (map[string]int64, error)
	GetOSCount(ctx context.Context, Id uuid.UUID) (map[string]int64, error)
	GetDeviceCount(ctx context.Context, Id uuid.UUID) (map[string]int64, error)
}
