package oms

import (
	"time"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

func (s *OMS) AddOrderToMap(order *model.Order) {
	s.orderIDMapping.Store(order.OrderID, order)
}

func (s *OMS) GetOrderByOrderID(orderID string) (*model.Order, error) {
	var order any
	var ok bool
	if order, ok = s.orderIDMapping.Load(orderID); !ok {
		return nil, errOrderIDNotFound
	}

	return order.(*model.Order), nil
}

func (s *OMS) DeleteOrderByOrderID(orderID string) {
	s.orderIDMapping.Delete(orderID)
}

func (s *OMS) startCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopCh:
			return
		}
	}
}

func (s *OMS) cleanup() {
	s.orderIDMapping.Range(func(k, v any) bool {
		order := v.(*model.Order)
		if order.IsEnd() {
			s.DeleteOrderByOrderID(order.OrderID)
		}
		return true
	})
}
