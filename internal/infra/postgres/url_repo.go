package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/labib0x9/short/internal/domain/url"
)

type urlRepo struct {
	db *sqlx.DB
}

func NewUrlRepository(db *sqlx.DB) url.UrlRepository {
	return &urlRepo{
		db: db,
	}
}

func (u *urlRepo) Create(url url.Url) error                          {}
func (u *urlRepo) GetByShortCode(shortCode string) (*url.Url, error) {}



package repositories

// import (
// 	"database/sql"
// 	"log/slog"
// 	"urlshortener/models"
// 	"urlshortener/utils"
// )

// // MysqlUrlRepository implements UrlRepository using a MySQL database as the backend
// // It provides methods to create and retrieve URL mappings from the MySQL database
// type MysqlUrlRepository struct {
// 	db *sql.DB // Database connection
// }

// // NewMysqlUrlRepository creates a new MysqlUrlRepository with the given database connection
// func NewMysqlUrlRepository(db *sql.DB) *MysqlUrlRepository {
// 	return &MysqlUrlRepository{
// 		db: db,
// 	}
// }

// // Create inserts a new URL mapping into the MySQL database
// func (u *MysqlUrlRepository) Create(url models.Url) error {
// 	query := "INSERT INTO urls (url, short_url, created_at, expire) VALUES (?, ?, ?, ?)"
// 	_, err := u.db.Exec(query, url.URL, url.ShortURL, url.CreatedAt, url.Expire)
// 	if err != nil {
// 		slog.Error(" [mysql_url_repository.go] [URL INSERT] ", slog.Any("error", err))
// 		return utils.ErrDatabaseInsert
// 	}
// 	return nil
// }

// // GetByShortCode retrieves a URL mapping by its short code from the MySQL database
// func (u *MysqlUrlRepository) GetByShortCode(shortCode string) (*models.Url, error) {
// 	query := "SELECT id, url, short_url, created_at, expire FROM urls WHERE short_url = ?"
// 	row := u.db.QueryRow(query, shortCode)
// 	var url models.Url
// 	err := row.Scan(&url.Id, &url.URL, &url.ShortURL, &url.CreatedAt, &url.Expire)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, nil // No result found
// 		}
// 		slog.Error(" [mysql_url_repository.go] [URL QUERY] ", slog.Any("error", err))
// 		return nil, utils.ErrDatabaseQuery
// 	}
// 	return &url, nil
// }
