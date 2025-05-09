package transactor

import (
	"context"

	"gorm.io/gorm"
)

// UnitOfWork 封装事务行为
type UnitOfWork interface {
	RunTransactional(ctx context.Context, fn func(ctx context.Context) error) error
}

type txKey struct{}

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetTx(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		return nil
	}
	return tx
}
