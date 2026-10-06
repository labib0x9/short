package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labib0x9/short/internal/port/db"
)

type pgxResult struct {
	tag pgconn.CommandTag
}

func (r pgxResult) RowsAffected() int64 {
	return r.tag.RowsAffected()
}

type pgxPoolAdapter struct {
	pool *pgxpool.Pool
}

func NewPgxPoolAdapter(pool *pgxpool.Pool) db.Operator {
	return &pgxPoolAdapter{pool: pool}
}

func (a *pgxPoolAdapter) Exec(ctx context.Context, query string, args ...any) (db.Result, error) {
	tag, err := a.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgxResult{tag: tag}, nil
}

func (a *pgxPoolAdapter) Query(ctx context.Context, query string, args ...any) (db.Rows, error) {
	return a.pool.Query(ctx, query, args...)
}

func (a *pgxPoolAdapter) QueryRow(ctx context.Context, query string, args ...any) db.Row {
	return a.pool.QueryRow(ctx, query, args...)
}

func (a *pgxPoolAdapter) Close(ctx context.Context) error {
	a.pool.Close()
	return nil
}

type pgxConnAdapter struct {
	conn *pgx.Conn
}

func NewPgxConnAdapter(conn *pgx.Conn) db.Operator {
	return &pgxConnAdapter{conn: conn}
}

func (a *pgxConnAdapter) Exec(ctx context.Context, query string, args ...any) (db.Result, error) {
	tag, err := a.conn.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgxResult{tag: tag}, nil
}

func (a *pgxConnAdapter) Query(ctx context.Context, query string, args ...any) (db.Rows, error) {
	return a.conn.Query(ctx, query, args...)
}

func (a *pgxConnAdapter) QueryRow(ctx context.Context, query string, args ...any) db.Row {
	return a.conn.QueryRow(ctx, query, args...)
}

func (a *pgxConnAdapter) Close(ctx context.Context) error {
	return a.conn.Close(ctx)
}

type pgxTxAdapter struct {
	tx pgx.Tx
}

func NewPgxTxAdapter(tx pgx.Tx) db.Operator {
	return &pgxTxAdapter{tx: tx}
}

func (a *pgxTxAdapter) Exec(ctx context.Context, query string, args ...any) (db.Result, error) {
	tag, err := a.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgxResult{tag: tag}, nil
}

func (a *pgxTxAdapter) Query(ctx context.Context, query string, args ...any) (db.Rows, error) {
	return a.tx.Query(ctx, query, args...)
}

func (a *pgxTxAdapter) QueryRow(ctx context.Context, query string, args ...any) db.Row {
	return a.tx.QueryRow(ctx, query, args...)
}

func (a *pgxTxAdapter) Close(ctx context.Context) error {
	return a.tx.Rollback(ctx)
}

// NewPgxAdapter automatically returns a db.Operator for *pgxpool.Pool, *pgx.Conn, or pgx.Tx
func NewPgxAdapter(v any) db.Operator {
	switch val := v.(type) {
	case *pgxpool.Pool:
		return NewPgxPoolAdapter(val)
	case *pgx.Conn:
		return NewPgxConnAdapter(val)
	case pgx.Tx:
		return NewPgxTxAdapter(val)
	case db.Operator:
		return val
	default:
		panic("unsupported pgx type for adapter")
	}
}

