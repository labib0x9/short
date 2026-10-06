package postgres

import (
	"context"
	"log/slog"
	"net"
	neturl "net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labib0x9/short/config"
	"github.com/labib0x9/short/internal/port/db"
)

func newConnectionString(cfg *config.PostgreSQL) string {
	var user *neturl.Userinfo
	if cfg.Pass != "" {
		user = neturl.UserPassword(cfg.User, cfg.Pass)
	} else if cfg.User != "" {
		user = neturl.User(cfg.User)
	}

	u := &neturl.URL{
		Scheme: "postgres",
		User:   user,
		Host:   net.JoinHostPort(cfg.Addr, cfg.Port),
		Path:   strings.TrimPrefix(cfg.DatabaseName, "/"),
	}
	if cfg.SslMode != "" {
		q := u.Query()
		q.Set("sslmode", cfg.SslMode)
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func newSuperConnectionString(cfg *config.PostgreSQL) string {
	var user *neturl.Userinfo
	if cfg.SuperUser != "" {
		user = neturl.User(cfg.SuperUser)
	}

	u := &neturl.URL{
		Scheme: "postgres",
		User:   user,
		Host:   net.JoinHostPort(cfg.Addr, cfg.Port),
		Path:   strings.TrimPrefix(cfg.SuperDatabase, "/"),
	}
	if cfg.SslMode != "" {
		q := u.Query()
		q.Set("sslmode", cfg.SslMode)
		u.RawQuery = q.Encode()
	}
	return u.String()
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
