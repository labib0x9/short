package db

import "context"

type TxManager interface {
	With(ctx context.Context, fn func(ctx context.Context) error) error
}
