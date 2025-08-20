package repo

import (
	"context"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"gorm.io/gorm"
)

type OrderEventSQLRepo struct {
	db *gorm.DB
}

func NewOrderEventSQLRepo(db *gorm.DB) *OrderEventSQLRepo {
	return &OrderEventSQLRepo{
		db: db,
	}
}

func (s *OrderEventSQLRepo) dbWithContext(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

func (r *OrderEventSQLRepo) Create(ctx context.Context, record *model.OrderEvent) (*model.OrderEvent, error) {
	return record, r.dbWithContext(ctx).Create(record).Error
}

func (r *OrderEventSQLRepo) BulkCreate(ctx context.Context, records []*model.OrderEvent) ([]*model.OrderEvent, error) {
	return records, r.dbWithContext(ctx).Create(records).Error

}
