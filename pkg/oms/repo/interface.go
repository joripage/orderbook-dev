package repo

import (
	"context"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type IOrder interface{}

type IOrderEvent interface {
	Create(ctx context.Context, record *model.OrderEvent) (*model.OrderEvent, error)
	BulkCreate(ctx context.Context, records []*model.OrderEvent) ([]*model.OrderEvent, error)
}
