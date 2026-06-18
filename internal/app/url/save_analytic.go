package url

import (
	"context"

	"github.com/labib0x9/short/internal/domain/queue"
	"github.com/labib0x9/short/internal/domain/url"
	"github.com/mileusna/useragent"
)

// to-do
// Get country from ip
func (s *service) Save(ctx context.Context, msg queue.ClickEvent) error {
	device, browser, os := parseUserAgent(msg.UserAgent)
	if device == "" || browser == "" || os == "" {
		// should not be empty
	}

	return s.txMngr.With(ctx, func(ctx context.Context) error {
		found, err := s.urlRepo.GetByShortCode(ctx, msg.ShortCode)
		if err != nil {
			return err
		}

		click := url.Click{
			UrlId:      found.Id,
			Referer:    msg.Referer,
			Country:    "",
			DeviceType: device,
			Os:         os,
			Browser:    browser,
			ClickedAt:  msg.ClickedAt,
		}

		err = s.analysisRepo.Create(ctx, click)
		if err != nil {
			return err
		}

		return s.urlRepo.Update(ctx, found.Id, msg.ClickedAt)
	})
}

func parseUserAgent(agent string) (device string, browser string, os string) {
	ua := useragent.Parse(agent)

	switch {
	case ua.Mobile:
		device = "Mobile"
	case ua.Desktop:
		device = "Desktop"
	}

	os = ua.OS

	browser = ua.Name

	switch {
	case device == "":
		device = "unknown"
		fallthrough
	case browser == "":
		browser = "unknown"
		fallthrough
	case os == "":
		os = "unknown"
	}

	return
}
