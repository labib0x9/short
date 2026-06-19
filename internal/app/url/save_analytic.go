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

	_, err := s.txMngr.With(ctx, func(ctx context.Context) (any, error) {
		found, err := s.urlRepo.GetByShortCode(ctx, msg.ShortCode)
		if err != nil {
			return nil, err
		}

		if found.IsExpired() {
			return nil, url.ErrShortCodeExpired
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

		click.SetEmptyFields()

		err = s.analysisRepo.Create(ctx, click)
		if err != nil {
			return nil, err
		}

		return nil, s.urlRepo.Update(ctx, found.Id, msg.ClickedAt)
	})
	return err
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
	return
}
