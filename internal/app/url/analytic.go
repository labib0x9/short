package url

import (
	"context"

	"github.com/labib0x9/short/internal/domain/url"
)

func (s *service) Analysis(ctx context.Context, code string) (url.Analysis, error) {
	result, err := s.txMngr.With(ctx, func(ctx context.Context) (any, error) {
		foundUrl, err := s.urlRepo.GetMetadata(ctx, code)
		if err != nil {
			return nil, err
		}

		if foundUrl.IsExpired() {
			return nil, url.ErrShortCodeExpired
		}

		browser, err := s.analysisRepo.GetBrowserCount(ctx, foundUrl.Id)
		if err != nil {
			return nil, err
		}

		device, err := s.analysisRepo.GetDeviceCount(ctx, foundUrl.Id)
		if err != nil {
			return nil, err
		}

		os, err := s.analysisRepo.GetOSCount(ctx, foundUrl.Id)
		if err != nil {
			return nil, err
		}
		return &url.Analysis{
			ShortURL:   code,
			ClickCount: foundUrl.ClickCount,
			Browser:    browser,
			Device:     device,
			OS:         os,
			ExpireAt:   foundUrl.ExpireAt,
		}, nil
	})

	if err != nil {
		return url.Analysis{}, err
	}

	if analytics, ok := result.(*url.Analysis); ok {
		return *analytics, nil
	}

	return url.Analysis{}, url.ErrGetAnalyticsFailed
}
