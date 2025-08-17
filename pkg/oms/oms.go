package oms

import (
	"context"
	"log"
	"sync"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/joripage/orderbook-dev/pkg/orderbook"
)

type OMS struct {
	orderGateway     OrderGateway
	orderbookManager *orderbook.OrderBookManager

	orderMapping sync.Map
}

var totalMatchQty int64 = 0
var totalMatchCount int64 = 0

func NewOMS(orderGateway OrderGateway) *OMS {
	orderbookManager := orderbook.NewOrderBookManager(&orderbook.OrderBookManagerConfig{
		EnableIceberg: true,
	})

	oms := &OMS{
		orderGateway:     orderGateway,
		orderbookManager: orderbookManager,
	}

	cb := func(results []orderbook.MatchResult) {
		for _, r := range results {
			log.Printf("Match: BUY[%s] <=> SELL[%s] @ %.2f Qty %d\n",
				r.OrderID, r.CounterOrderID, r.Price, r.Qty)

			totalMatchQty += r.Qty
			totalMatchCount += 1
			log.Printf("TotalMatchCount: %d, TotalMatchQty: %d\n", totalMatchCount, totalMatchQty)

			order, err := oms.GetOrderByOrderID(r.OrderID)
			if err != nil {
				log.Printf("match orderID=%s not found", r.OrderID)
				continue
			}

			order.UpdateMatchResult(&r)
			oms.orderGateway.OnOrderReport(context.Background(), order)

			counterOrder, err := oms.GetOrderByOrderID(r.CounterOrderID)
			if err != nil {
				log.Printf("match counterOrderID=%s not found", r.CounterOrderID)
				continue
			}

			counterOrder.UpdateMatchResult(&r)
			oms.orderGateway.OnOrderReport(context.Background(), counterOrder)
		}
	}
	oms.orderbookManager.RegisterTradeCallback(cb)

	return oms
}

func (s *OMS) Start(ctx context.Context) {
	s.orderGateway.Start(ctx)
}

func (s *OMS) AddOrder(ctx context.Context, addOrder *model.AddOrder) {
	order := &model.Order{}
	order.UpdateAddOrder(addOrder)
	s.AddOrderToMap(order)

	// report pending new
	s.orderGateway.OnOrderReport(ctx, order)
	s.orderbookManager.AddOrder(&orderbook.Order{
		ID:          order.OrderID,
		Symbol:      order.Symbol,
		Side:        orderbook.Side(order.Side),
		Price:       order.Price,
		Qty:         order.Quantity,
		Type:        orderbook.OrderType(order.Type),
		TimeInForce: orderbook.TimeInForce(order.TimeInForce),
	})

	// book success -> change pending new to new
	order.Status = model.OrderStatusNew
	s.orderGateway.OnOrderReport(ctx, order)
}
