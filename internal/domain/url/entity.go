package url

import (
	"time"

	"github.com/google/uuid"
)

type Url struct {
	Id        uuid.UUID `json:"id" db:"id"`
	URL       string    `json:"url" db:"url"`
	ShortURL  string    `json:"short" db:"short"`
	CreatedAt time.Time `json:"created_at"    db:"created_at"`
	ExpireAt  time.Time `json:"expire_at"    db:"expire_at"`
}

type Click struct {
	Id         uuid.UUID `json:"id" db:"id"`
	Referer    string    `json:"referer"    db:"referer"`
	Country    string    `json:"country"    db:"country"`
	DeviceType string    `json:"device"    db:"device"`
	ClickedAt  time.Time `json:"clicked_at"    db:"clicked_at"`
}

type Stat struct {
	Id            uuid.UUID `json:"id" db:"id"`
	Total         int64     `json:"total" db:"total"`
	LastClickedAt time.Time `json:"last_clicked_at"    db:"last_clicked_at"`
}
