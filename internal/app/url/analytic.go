package url

import "github.com/labib0x9/short/internal/domain/url"

func (s *service) Analysis(code string) (url.Analysis, error) {

	foundUrl, err := s.urlRepo.GetMetadata(code)
	if err != nil {
		return url.Analysis{}, err
	}

	browser, err := s.analysisRepo.GetBrowserCount(foundUrl.Id)
	if err != nil {
		return url.Analysis{}, err
	}

	device, err := s.analysisRepo.GetDeviceCount(foundUrl.Id)
	if err != nil {
		return url.Analysis{}, err
	}

	os, err := s.analysisRepo.GetOSCount(foundUrl.Id)
	if err != nil {
		return url.Analysis{}, err
	}

	return url.Analysis{
		ShortURL:   code,
		ClickCount: foundUrl.ClickCount,
		Browser:    browser,
		Device:     device,
		OS:         os,
		ExpireAt:   foundUrl.ExpireAt,
	}, nil
}
