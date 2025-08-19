package repo

import (
	"context"

	"gorm.io/gorm"
)

type OrderSQLRepo struct {
	db *gorm.DB
}

func NewOrderSQLRepo(db *gorm.DB) *OrderSQLRepo {
	return &OrderSQLRepo{
		db: db,
	}
}

func (s *OrderSQLRepo) dbWithContext(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}
