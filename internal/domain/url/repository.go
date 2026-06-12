package url

import (
	"time"

	"github.com/google/uuid"
)

type UrlRepository interface {
	Create(url Url) error
	GetByShortCode(shortCode string) (*Url, error)
	Update(id uuid.UUID, lastClickedAt time.Time) error
	GetMetadata(code string) (*Url, error)
}

type AnalyticsRepository interface {
	Create(click Click) error
	GetBrowserCount(Id uuid.UUID) (map[string]int64, error)
	GetOSCount(Id uuid.UUID) (map[string]int64, error)
	GetDeviceCount(Id uuid.UUID) (map[string]int64, error)
}
