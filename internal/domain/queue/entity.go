package queue

import "time"

type ClickEvent struct {
	ShortCode string    `json:"short_code"`
	ClickedAt time.Time `json:"clicked_at"`
	Referer   string    `json:"referer"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	Retries   int       `json:"retires"`
}
