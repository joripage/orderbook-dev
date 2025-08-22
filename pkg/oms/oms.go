package oms

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
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

	orderIDMapping sync.Map
	stopCh         chan struct{}
	// gatewayIDMapping sync.Map

	rules []riskrule.RiskRule
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
		stopCh:           make(chan struct{}),
	}
	// go oms.startCleaner(10 * time.Second)

	return oms
}

func (s *OMS) Start(ctx context.Context) {
	s.orderGateway.Start(ctx)
}

func (s *OMS) Stop() {
	close(s.stopCh)
}

func (s *OMS) AddOrder(ctx context.Context, addOrder *model.AddOrder) error {
	// todo: check riskrule
	orderID := s.eventstore.GetLatestGatewayID(addOrder.GatewayID)
	if orderID != "" {
		return errDuplicateOrder
	}

	order := &model.Order{}
	order.UpdateAddOrder(addOrder)
	s.AddOrderToMap(order)

	results := s.orderbookManager.AddOrder(&orderbook.Order{
		ID:          order.OrderID,
		Symbol:      order.Symbol,
		Side:        orderbook.Side(order.Side),
		Price:       order.Price,
		Qty:         order.Quantity,
		Type:        orderbook.OrderType(order.Type),
		TimeInForce: orderbook.TimeInForce(order.TimeInForce),
	})

	// book success -> change pending new to new
	bkOrder := *order
	now := time.Now()
	// ov, fnReset := model.NewOrderEventUsingPool(bkOrder, now)
	// s.eventstore.AddEvent(ov)
	// fnReset()
	s.eventstore.AddEvent(model.NewOrderEvent(bkOrder, now))
	s.orderGateway.OnOrderReport(ctx, bkOrder)

	s.processMatchResult(results)

	return nil
}

func (s *OMS) CancelOrder(ctx context.Context, cancelOrder *model.CancelOrder) error {
	// todo: check riskrule
	orderID := s.eventstore.GetOrderID(cancelOrder.OrigGatewayID)
	order, err := s.GetOrderByOrderID(orderID)
	if err != nil {
		return errGatewayIDNotFound
	}

	if !order.CanCancel() {
		return errInvalidOrderStatus
	}

	err = s.orderbookManager.CancelOrder(order.Symbol, order.OrderID)
	_ = err
	order.UpdateCancelOrder(cancelOrder)

	bkOrder := *order
	now := time.Now()
	// ov, fnReset := model.NewOrderEventUsingPool(bkOrder, now)
	// s.eventstore.AddEvent(ov)
	// fnReset()
	s.eventstore.AddEvent(model.NewOrderEvent(bkOrder, now))
	s.orderGateway.OnOrderReport(ctx, bkOrder)

	return nil
}

func (s *OMS) ModifyOrder(ctx context.Context, modifyOrder *model.ModifyOrder) error {
	// todo: check riskrule
	orderID := s.eventstore.GetOrderID(modifyOrder.OrigGatewayID)
	order, err := s.GetOrderByOrderID(orderID)
	if err != nil {
		return errGatewayIDNotFound
	}

	if !order.CanModify() {
		return errInvalidOrderStatus
	}

	newPrice, newQty := modifyOrder.NewPrice.InexactFloat64(), modifyOrder.NewQuantity.IntPart()
	results, err := s.orderbookManager.ModifyOrder(order.Symbol, order.OrderID, newPrice, newQty)
	_ = err
	order.UpdateModifyOrder(modifyOrder)

	bkOrder := *order
	now := time.Now()
	// ov, fnReset := model.NewOrderEventUsingPool(bkOrder, now)
	// s.eventstore.AddEvent(ov)
	// fnReset()
	s.eventstore.AddEvent(model.NewOrderEvent(bkOrder, now))
	s.orderGateway.OnOrderReport(ctx, bkOrder)

	s.processMatchResult(results)

	return nil
}

func (s *OMS) processMatchResult(results []*orderbook.MatchResult) {
	for _, r := range results {
		// log.Printf("Match: BUY[%s] <=> SELL[%s] @ %.2f Qty %d\n",
		// 	r.OrderID, r.CounterOrderID, r.Price, r.Qty)

		atomic.AddInt64(&totalMatchQty, r.Qty)
		atomic.AddInt64(&totalMatchCount, 1)
		if totalMatchCount%10000 == 0 {
			log.Printf("TotalMatchCount: %d, TotalMatchQty: %d\n", totalMatchCount, totalMatchQty)
		}

		order, err := s.GetOrderByOrderID(r.OrderID)
		if err != nil {
			log.Printf("match orderID=%s not found", r.OrderID)
			continue
		}

		order.UpdateMatchResult(r)
		bkOrder := *order
		now := time.Now()
		// ov, fnReset := model.NewOrderEventUsingPool(bkOrder, now)
		// s.eventstore.AddEvent(ov)
		// fnReset()
		s.eventstore.AddEvent(model.NewOrderEvent(bkOrder, now))
		s.orderGateway.OnOrderReport(context.Background(), bkOrder)

		counterOrder, err := s.GetOrderByOrderID(r.CounterOrderID)
		if err != nil {
			log.Printf("match counterOrderID=%s not found", r.CounterOrderID)
			continue
		}

		counterOrder.UpdateMatchResult(r)
		bkCounterOrder := *counterOrder
		// ovCounter, fnReset := model.NewOrderEventUsingPool(bkCounterOrder, now)
		// s.eventstore.AddEvent(ovCounter)
		// fnReset()
		s.eventstore.AddEvent(model.NewOrderEvent(bkCounterOrder, now))
		s.orderGateway.OnOrderReport(context.Background(), bkCounterOrder)
	}
}
