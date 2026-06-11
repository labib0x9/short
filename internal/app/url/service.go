package url

import "github.com/labib0x9/short/internal/domain/url"

type Service interface {
	// Shorten()
}

type service struct {
	urlRepo      url.UrlRepository
	analysisRepo url.AnalyticsRepository
}

func NewService(
	urlRepo url.UrlRepository,
	analysisRepo url.AnalyticsRepository,
) Service {
	return service{
		urlRepo:      urlRepo,
		analysisRepo: analysisRepo,
	}
}
