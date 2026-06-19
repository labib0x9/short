package postgres

import (
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/labib0x9/short/config"
	_ "github.com/lib/pq"
)

func NewPostgresConn(cfg *config.PostgreSQL) *sqlx.DB {
	dbSource := newConnectionString(cfg)
	conn, err := sqlx.Connect("postgres", dbSource)
	if err != nil {
		panic(err)
	}
	return conn
}

func newConnectionString(cfg *config.PostgreSQL) string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", cfg.User, cfg.Pass, cfg.Addr, cfg.Port, cfg.DatabaseName, cfg.SslMode)
}

func newSuperConnectionString(cfg *config.PostgreSQL) string {
	return fmt.Sprintf("user=%s password= host=%s port=%s dbname=%s sslmode=%s", cfg.SuperUser, cfg.Addr, cfg.Port, cfg.SuperDatabase, cfg.SslMode)
}

func NewPostgresSuperConn(cfg *config.PostgreSQL) *sqlx.DB {
	dbSource := newSuperConnectionString(cfg)
	conn, err := sqlx.Connect("postgres", dbSource)
	if err != nil {
		panic(err)
	}
	return conn
}
