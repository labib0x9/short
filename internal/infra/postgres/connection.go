package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/port/db"
)

func newConnectionString(cfg *config.PostgreSQL) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Pass,
		cfg.Addr,
		cfg.Port,
		cfg.DatabaseName,
		cfg.SslMode,
	)
}

func newSuperConnectionString(cfg *config.PostgreSQL) string {
	return fmt.Sprintf(
		"postgres://%s@%s:%s/%s?sslmode=%s",
		cfg.SuperUser,
		cfg.Addr,
		cfg.Port,
		cfg.SuperDatabase,
		cfg.SslMode,
	)
}

func NewPostgresPool(ctx context.Context, cfg *config.PostgreSQL) *pgxpool.Pool {
	dbSource := newConnectionString(cfg)
	poolConfig, err := pgxpool.ParseConfig(dbSource)
	if err != nil {
		panic(err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 10
	poolConfig.MinIdleConns = 15
	poolConfig.MaxConnIdleTime = 20 * time.Minute

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	connPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		panic(err)
	}
	slog.Info("Postgres connected")
	return connPool
}

func NewPostgresConn(ctx context.Context, cfg *config.PostgreSQL) db.Operator {
	connPool := NewPostgresPool(ctx, cfg)
	return NewPgxAdapter(connPool)
}

func NewPostgresSuperConn(ctx context.Context, cfg *config.PostgreSQL) db.Operator {
	dbSource := newSuperConnectionString(cfg)
	conn, err := pgx.Connect(ctx, dbSource)
	if err != nil {
		panic(err)
	}
	return NewPgxAdapter(conn)
}
