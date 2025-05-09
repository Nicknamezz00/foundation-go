package transactor

import (
	"context"
	"gorm.io/gorm"
)

type GormUnitOfWork struct {
	db *gorm.DB
}

func NewGormUnitOfWork(db *gorm.DB) *GormUnitOfWork {
	return &GormUnitOfWork{db: db}
}

func (u *GormUnitOfWork) RunTransactional(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.DoTx(ctx, u.db, fn)
}

func (u *GormUnitOfWork) DoTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	ctxWithTx := WithTx(ctx, tx)
	if err := fn(ctxWithTx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
