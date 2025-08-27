package oms

import (
	"fmt"
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
	fmt.Println("clean up")
	// var stats runtime.MemStats
	// runtime.ReadMemStats(&stats)
	// fmt.Printf("HeapObjects: %d, NumGC: %v, LastPause: %vµs\n",
	// 	stats.HeapObjects,
	// 	stats.NumGC,
	// 	stats.PauseNs[(stats.NumGC+255)%256]/1000,
	// )
	s.orderIDMapping.Range(func(k, v any) bool {
		order := v.(*model.Order)
		if order.IsEnd() {
			s.DeleteOrderByOrderID(order.OrderID)
			s.eventstore.DeleteChainByOrderID(order.OrderID)
		}
		return true
	})
}
