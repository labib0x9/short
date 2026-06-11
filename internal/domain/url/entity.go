package url

import (
	"time"

	"github.com/google/uuid"
)

type Url struct {
	Id        uuid.UUID
	URL       string
	ShortURL  string
	CreatedAt time.Time
	Expire    time.Time
}

type Click struct {
	Id         uuid.UUID
	ClickedAt  time.Time
	Referer    string
	Country    string
	DeviceType string
}

type Stat struct {
	Id            uuid.UUID
	Total         int64
	LastClickedAt time.Time
}
