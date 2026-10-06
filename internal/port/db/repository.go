package db

import "context"

type TxManager interface {
	With(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)
}

type Result interface {
	RowsAffected() int64
}

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Operator interface {
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	Close(ctx context.Context) error
}
