package url

import (
	"time"

	"github.com/google/uuid"
)

type Url struct {
	Id            uuid.UUID  `json:"id" db:"id"`
	URL           string     `json:"url" db:"url"`
	ShortURL      string     `json:"short" db:"short"`
	ClickCount    int64      `json:"total" db:"total"`
	LastClickedAt *time.Time `json:"last_clicked_at"    db:"last_clicked_at"`
	CreatedAt     time.Time  `json:"created_at"    db:"created_at"`
	ExpireAt      *time.Time `json:"expire_at"    db:"expire_at"`
}

type Click struct {
	Id         int64     `json:"id" db:"id"`
	UrlId      uuid.UUID `json:"url_id" db:"url_id"`
	Referer    string    `json:"referer"    db:"referer"`
	Country    string    `json:"country"    db:"country"`
	DeviceType string    `json:"device"    db:"device"`
	Os         string    `json:"os"    db:"os"`
	Browser    string    `json:"browser"    db:"browser"`
	ClickedAt  time.Time `json:"clicked_at"    db:"clicked_at"`
}

type Analysis struct {
	ShortURL   string           `json:"short"`
	ClickCount int64            `json:"total_count"`
	Browser    map[string]int64 `json:"browser"`
	Device     map[string]int64 `json:"device"`
	OS         map[string]int64 `json:"os"`
	ExpireAt   *time.Time       `json:"expire_at"`
}
