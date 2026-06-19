package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labib0x9/short/internal/domain/url"
)

// every operation can run with transaction
// if there is no transaction, fallback to default db

type urlRepo struct {
	db *sqlx.DB
}

func NewUrlRepository(db *sqlx.DB) url.UrlRepository {
	return &urlRepo{
		db: db,
	}
}

func (u *urlRepo) Create(ctx context.Context, url url.Url) error {
	db := getDBFromCtx(ctx, u.db)
	query := `insert into 
		urls(url, short, expire_at)
		values(:url, :short, :expire_at)
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, url)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func (u *urlRepo) GetByShortCode(ctx context.Context, shortCode string) (*url.Url, error) {
	db := getDBFromCtx(ctx, u.db)
	query := `select * from urls where short = $1`
	var found url.Url
	if err := sqlx.GetContext(ctx, db, &found, query, shortCode); err != nil {
		return nil, err
	}
	return &found, nil
}

func (u *urlRepo) Update(ctx context.Context, id uuid.UUID, lastClickedAt time.Time) error {
	db := getDBFromCtx(ctx, u.db)
	query := `
		update urls
		set
			last_clicked_at = $1,
			total = COALESCE(total, 0) + 1
		where id = $2`
	_, err := db.ExecContext(ctx, query, lastClickedAt, id)
	return err
}

func (u *urlRepo) GetMetadata(ctx context.Context, code string) (*url.Url, error) {
	db := getDBFromCtx(ctx, u.db)
	query := `
		select
			id, total, last_clicked_at, created_at, expire_at
		from urls
		where short = $1`
	var found url.Url
	if err := sqlx.GetContext(ctx, db, &found, query, code); err != nil {
		return nil, err
	}
	return &found, nil
}

func (u *urlRepo) DeleteByExpireAt(ctx context.Context) error {
	db := getDBFromCtx(ctx, u.db)
	query := `
		delete from urls where expire_at is not NULL and expire_at < NOW()
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

type analysisRepo struct {
	db *sqlx.DB
}

func NewAnalysisRepository(db *sqlx.DB) url.AnalyticsRepository {
	return &analysisRepo{
		db: db,
	}
}

func (a *analysisRepo) Create(ctx context.Context, click url.Click) error {
	db := getDBFromCtx(ctx, a.db)
	query := `insert into 
		clicks(url_id, referer, country, device, os, browser, clicked_at)
		values(:url_id, :referer, :country, :device, :os, :browser, :clicked_at)
	`

	// db.ExecContext()
	rows, err := sqlx.NamedQueryContext(ctx, db, query, click)

	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func (a *analysisRepo) GetBrowserCount(ctx context.Context, Id uuid.UUID) (map[string]int64, error) {
	db := getDBFromCtx(ctx, a.db)
	result := map[string]int64{}
	query := `
		select
			browser, count(*)
		from
			clicks
		where url_id = $1
		group by browser
	`
	rows, err := db.QueryContext(ctx, query, Id)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		var k string
		var v int64
		rows.Scan(&k, &v)
		result[k] = v
	}
	return result, nil
}

func (a *analysisRepo) GetDeviceCount(ctx context.Context, Id uuid.UUID) (map[string]int64, error) {
	db := getDBFromCtx(ctx, a.db)
	result := map[string]int64{}
	query := `
		select
			device, count(*)
		from
			clicks
		where url_id = $1
		group by device
	`
	rows, err := db.QueryContext(ctx, query, Id)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		var k string
		var v int64
		rows.Scan(&k, &v)
		result[k] = v
	}
	return result, nil
}

func (a *analysisRepo) GetOSCount(ctx context.Context, Id uuid.UUID) (map[string]int64, error) {
	db := getDBFromCtx(ctx, a.db)
	result := map[string]int64{}
	query := `
		select
			os, count(*)
		from
			clicks
		where url_id = $1
		group by os
	`
	rows, err := db.QueryContext(ctx, query, Id)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		var k string
		var v int64
		rows.Scan(&k, &v)
		result[k] = v
	}
	return result, nil
}
