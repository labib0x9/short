package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labib0x9/short/internal/port/db"
)

type txKey struct{}

type txManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) db.TxManager {
	return &txManager{
		pool: pool,
	}
}

func (t *txManager) With(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return nil, err
	}

	txOp := NewPgxAdapter(tx)
	txCtx := context.WithValue(ctx, txKey{}, txOp)

	result, err := fn(txCtx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	return result, tx.Commit(ctx)
}

// getDBFromCtx extracts the transactional db.Operator from context, or falls back to defaultDB
func getDBFromCtx(ctx context.Context, defaultDB db.Operator) db.Operator {
	if txOp, ok := ctx.Value(txKey{}).(db.Operator); ok {
		return txOp
	}
	return defaultDB
}

