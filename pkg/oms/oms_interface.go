package oms

import (
	"context"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type IOMS interface {
	AddOrder(ctx context.Context, addOrder *model.AddOrder) error
	ModifyOrder(ctx context.Context, gatewayID, origGatewayID string, newPrice float64, newQty int64) error
	CancelOrder(ctx context.Context, gatewayID string) error
}
