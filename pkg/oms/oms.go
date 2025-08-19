package oms

import (
	"context"
	"log"
	"sync"
	"time"

	eventstore "github.com/joripage/orderbook-dev/pkg/oms/event_store"
	"github.com/joripage/orderbook-dev/pkg/oms/model"
	riskrule "github.com/joripage/orderbook-dev/pkg/oms/risk_rule"
	"github.com/joripage/orderbook-dev/pkg/orderbook"
)

type OMS struct {
	orderGateway     OrderGateway
	orderbookManager *orderbook.OrderBookManager
	eventstore       eventstore.EventStore

	orderIDMapping   sync.Map
	gatewayIDMapping sync.Map
	rules            []riskrule.RiskRule
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
		eventstore:       eventstore.NewInMemoryEventStore(),
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

func (s *OMS) AddOrder(ctx context.Context, addOrder *model.AddOrder) error {
	// todo: check riskrule
	oldOrder, _ := s.GetOrderByGatewayID(addOrder.ID)
	if oldOrder != nil {
		return errDuplicateOrder
	}

	order := &model.Order{}
	order.UpdateAddOrder(addOrder)
	s.AddOrderToMap(order)

	// report pending new
	// s.orderGateway.OnOrderReport(ctx, order)
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
	s.eventstore.AddEvent(model.NewOrderEventNewOrder(order.OrderID, order.GatewayID, time.Now()))
	s.orderGateway.OnOrderReport(ctx, order)

	return nil
}

func (s *OMS) CancelOrder(ctx context.Context, gatewayID string) error {
	// todo: check riskrule
	order, err := s.GetOrderByGatewayID(gatewayID)
	if err != nil {
		return errGatewayIDNotFound
	}

	if !order.CanCancel() {
		return errInvalidOrderStatus
	}

	err = s.orderbookManager.CancelOrder(order.Symbol, order.OrderID)
	_ = err
	order.Status = model.OrderStatusCanceled
	s.eventstore.AddEvent(model.NewOrderEventCancel(order.OrderID, order.GatewayID, order.OrigGatewayID, time.Now()))
	s.orderGateway.OnOrderReport(ctx, order)

	return nil
}

func (s *OMS) ModifyOrder(ctx context.Context, gatewayID, origGatewayID string, newPrice float64, newQty int64) error {
	// todo: check riskrule
	order, err := s.GetOrderByGatewayID(origGatewayID)
	if err != nil {
		return errGatewayIDNotFound
	}

	if !order.CanModify() {
		return errInvalidOrderStatus
	}

	err = s.orderbookManager.ModifyOrder(order.Symbol, order.OrderID, newPrice, newQty)
	_ = err
	order.Status = model.OrderStatusReplaced
	s.eventstore.AddEvent(model.NewOrderEventCancelReplace(order.OrderID, order.GatewayID, order.OrigGatewayID, newPrice, newQty, time.Now()))
	s.orderGateway.OnOrderReport(ctx, order)

	return nil
}
