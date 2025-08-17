package oms

import (
	"errors"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

func (s *OMS) AddOrderToMap(order *model.Order) {
	s.orderMapping.Store(order.OrderID, order)
}

func (s *OMS) GetOrderByOrderID(orderID string) (*model.Order, error) {
	var order any
	var ok bool
	if order, ok = s.orderMapping.Load(orderID); !ok {
		return nil, errors.New("orderID not found")
	}

	return order.(*model.Order), nil
}
