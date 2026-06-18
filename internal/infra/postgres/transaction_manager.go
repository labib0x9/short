package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/labib0x9/short/internal/domain/db"
)

type txKey struct{}

type txManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) db.TxManager {
	return &txManager{
		db: db,
	}
}

func (t *txManager) With(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Get db connection from context, if absent fallback to db
func getDBFromCtx(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}
