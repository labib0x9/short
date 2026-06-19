package url

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/labib0x9/short/internal/domain/url"
	"github.com/labib0x9/short/internal/utils"
)

type ShortenResult struct {
	Msg      string     `json:"msg"`
	Code     string     `json:"code"`
	ShortUrl string     `json:"short_url"`
	ExpireAt *time.Time `json:"expire_at"`
}

func (s *service) Shorten(ctx context.Context, longUrl string, expireAt *time.Time, userAgent string) (ShortenResult, error) {
	if len(longUrl) <= 25 {
		return ShortenResult{}, url.ErrUrlShortLenght
	}

	uniqueId := utils.UniqueId(userAgent)
	short := utils.GetShortUrl(longUrl + uniqueId)

	createdAt := time.Now()

	existShort, err := s.urlRepo.GetByShortCode(ctx, short)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ShortenResult{}, err
	}

	if existShort != nil {
		return ShortenResult{}, url.ErrShortCodeCollision
	}

	value := url.Url{
		URL:       longUrl,
		ShortURL:  short,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
	}

	err = s.urlRepo.Create(ctx, value)
	if err != nil {
		return ShortenResult{}, err
	}

	return ShortenResult{
		Msg:      "success",
		Code:     value.ShortURL,
		ShortUrl: s.cnf.Prefix + value.ShortURL,
		ExpireAt: value.ExpireAt,
	}, nil
}
