package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labib0x9/short/internal/domain/url"
	"github.com/labib0x9/short/internal/port/db"
)

// every operation can run with transaction
// if there is no transaction, fallback to default db

type urlRepo struct {
	db db.Operator
}

func NewUrlRepository(db db.Operator) url.UrlRepository {
	return &urlRepo{
		db: db,
	}
}

func (u *urlRepo) Create(ctx context.Context, item url.Url) error {
	op := getDBFromCtx(ctx, u.db)
	query := `INSERT INTO urls(url, short, expire_at) VALUES (@url, @short, @expire_at)`
	_, err := op.Exec(ctx, query, pgx.NamedArgs{
		"url":       item.URL,
		"short":     item.ShortURL,
		"expire_at": item.ExpireAt,
	})
	return err
}

func (u *urlRepo) GetByShortCode(ctx context.Context, shortCode string) (*url.Url, error) {
	op := getDBFromCtx(ctx, u.db)
	query := `SELECT id, url, short, COALESCE(total, 0), last_clicked_at, created_at, expire_at FROM urls WHERE short = $1`
	var found url.Url
	err := op.QueryRow(ctx, query, shortCode).Scan(
		&found.Id,
		&found.URL,
		&found.ShortURL,
		&found.ClickCount,
		&found.LastClickedAt,
		&found.CreatedAt,
		&found.ExpireAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &found, nil
}

func (u *urlRepo) Update(ctx context.Context, id uuid.UUID, lastClickedAt time.Time) error {
	op := getDBFromCtx(ctx, u.db)
	query := `
		UPDATE urls
		SET
			last_clicked_at = @last_clicked_at,
			total = COALESCE(total, 0) + 1
		WHERE id = @id`
	_, err := op.Exec(ctx, query, pgx.NamedArgs{
		"last_clicked_at": lastClickedAt,
		"id":              id,
	})
	return err
}

func (u *urlRepo) GetMetadata(ctx context.Context, code string) (*url.Url, error) {
	op := getDBFromCtx(ctx, u.db)
	query := `
		SELECT
			id, COALESCE(total, 0), last_clicked_at, created_at, expire_at
		FROM urls
		WHERE short = $1`
	var found url.Url
	err := op.QueryRow(ctx, query, code).Scan(
		&found.Id,
		&found.ClickCount,
		&found.LastClickedAt,
		&found.CreatedAt,
		&found.ExpireAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &found, nil
}

func (u *urlRepo) DeleteByExpireAt(ctx context.Context) error {
	op := getDBFromCtx(ctx, u.db)
	query := `DELETE FROM urls WHERE expire_at IS NOT NULL AND expire_at < NOW()`
	_, err := op.Exec(ctx, query)
	return err
}

type analysisRepo struct {
	db db.Operator
}

func NewAnalysisRepository(db db.Operator) url.AnalyticsRepository {
	return &analysisRepo{
		db: db,
	}
}

func (a *analysisRepo) Create(ctx context.Context, click url.Click) error {
	op := getDBFromCtx(ctx, a.db)
	query := `
		INSERT INTO clicks(url_id, referer, country, device, os, browser, clicked_at)
		VALUES (@url_id, @referer, @country, @device, @os, @browser, @clicked_at)
	`
	_, err := op.Exec(
		ctx,
		query,
		pgx.NamedArgs{
			"url_id":     click.UrlId,
			"referer":    click.Referer,
			"country":    click.Country,
			"device":     click.DeviceType,
			"os":         click.Os,
			"browser":    click.Browser,
			"clicked_at": click.ClickedAt,
		},
	)
	return err
}

func (a *analysisRepo) GetBrowserCount(ctx context.Context, id uuid.UUID) (map[string]int64, error) {
	op := getDBFromCtx(ctx, a.db)
	result := map[string]int64{}
	query := `
		SELECT
			browser, count(*)
		FROM
			clicks
		WHERE url_id = $1
		GROUP BY browser
	`
	rows, err := op.Query(ctx, query, id)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var k string
		var v int64
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

func (a *analysisRepo) GetDeviceCount(ctx context.Context, id uuid.UUID) (map[string]int64, error) {
	op := getDBFromCtx(ctx, a.db)
	result := map[string]int64{}
	query := `
		SELECT
			device, count(*)
		FROM
			clicks
		WHERE url_id = $1
		GROUP BY device
	`
	rows, err := op.Query(ctx, query, id)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var k string
		var v int64
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

func (a *analysisRepo) GetOSCount(ctx context.Context, id uuid.UUID) (map[string]int64, error) {
	op := getDBFromCtx(ctx, a.db)
	result := map[string]int64{}
	query := `
		SELECT
			os, count(*)
		FROM
			clicks
		WHERE url_id = $1
		GROUP BY os
	`
	rows, err := op.Query(ctx, query, id)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var k string
		var v int64
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

