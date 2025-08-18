package oms

import (
	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

func (s *OMS) AddOrderToMap(order *model.Order) {
	s.orderIDMapping.Store(order.OrderID, order)
	s.gatewayIDMapping.Store(order.GatewayID, order)
}

func (s *OMS) GetOrderByOrderID(orderID string) (*model.Order, error) {
	var order any
	var ok bool
	if order, ok = s.orderIDMapping.Load(orderID); !ok {
		return nil, errOrderIDNotFound
	}

	return order.(*model.Order), nil
}

func (s *OMS) GetOrderByGatewayID(gatewayID string) (*model.Order, error) {
	var order any
	var ok bool
	if order, ok = s.gatewayIDMapping.Load(gatewayID); !ok {
		return nil, errGatewayIDNotFound
	}

	return order.(*model.Order), nil
}
