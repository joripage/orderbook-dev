package oms

import (
	"context"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type IOMS interface {
	AddOrder(ctx context.Context, addOrder *model.AddOrder)
	// ModifyOrder(ctx context.Context)
	CancelOrder(ctx context.Context, orderID string)
}
