package model

import (
	"fmt"
	"time"
)

type OrderEvent struct {
	EventID       string
	OrderID       string
	GatewayID     string
	OrigGatewayID string
	ExecType      OrderExecType
	Qty           int64
	Price         float64
	Timestamp     time.Time
}

func NewOrderEventNewOrder(orderID, gatewayID string, ts time.Time) *OrderEvent {
	return &OrderEvent{
		EventID:   NewEventID(orderID, OrderStatusNew),
		OrderID:   orderID,
		GatewayID: gatewayID,
		ExecType:  ExecTypeNew,
		Timestamp: ts,
	}
}

func NewOrderEventCancel(orderID, gatewayID, origGatewayID string, ts time.Time) *OrderEvent {
	return &OrderEvent{
		EventID:       NewEventID(orderID, OrderStatusCanceled),
		OrderID:       orderID,
		GatewayID:     gatewayID,
		OrigGatewayID: origGatewayID,
		ExecType:      ExecTypeCanceled,
		Timestamp:     ts,
	}
}

func NewOrderEventCancelReplace(orderID, gatewayID, origGatewayID string, price float64, qty int64, ts time.Time) *OrderEvent {
	return &OrderEvent{
		EventID:       NewEventID(orderID, OrderStatusReplaced),
		OrderID:       orderID,
		GatewayID:     gatewayID,
		OrigGatewayID: origGatewayID,
		ExecType:      ExecTypeReplaced,
		Qty:           qty,
		Price:         price,
		Timestamp:     ts,
	}
}

func NewEventID(orderID string, status OrderStatus) string {
	return fmt.Sprintf("%s-%s", orderID, status)
}
