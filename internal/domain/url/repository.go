package url

type UrlRepository interface {
	Create(url Url) error
	GetByShortCode(shortCode string) (*Url, error)
}

type AnalyticsRepository interface {
}
