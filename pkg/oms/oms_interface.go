package oms

import (
	"context"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type IOMS interface {
	AddOrder(ctx context.Context, addOrder *model.AddOrder) error
	ModifyOrder(ctx context.Context, modifyOrder *model.ModifyOrder) error
	CancelOrder(ctx context.Context, gatewayID, origGatewayID string) error
}
