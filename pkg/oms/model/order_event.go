package model

import (
	"fmt"
	"time"
)

type OrderEvent struct {
	EventID     string
	OrderID     string
	ClOrdID     string
	OrigClOrdID string
	ExecType    OrderExecType
	Qty         int64
	Price       float64
	Timestamp   time.Time
}

func NewOrderEventNewOrder(orderID, clOrdID string, ts time.Time) *OrderEvent {
	return &OrderEvent{
		EventID:   NewEventID(orderID, OrderStatusNew),
		OrderID:   orderID,
		ClOrdID:   clOrdID,
		ExecType:  ExecTypeNew,
		Timestamp: ts,
	}
}

func NewOrderEventCancel(orderID, clOrdID, origClOrdID string, ts time.Time) *OrderEvent {
	return &OrderEvent{
		EventID:     NewEventID(orderID, OrderStatusCanceled),
		OrderID:     orderID,
		ClOrdID:     clOrdID,
		OrigClOrdID: origClOrdID,
		ExecType:    ExecTypeCanceled,
		Timestamp:   ts,
	}
}

func NewOrderEventCancelReplace(orderID, clOrdID, origClOrdID string, price float64, qty int64, ts time.Time) *OrderEvent {
	return &OrderEvent{
		EventID:     NewEventID(orderID, OrderStatusReplaced),
		OrderID:     orderID,
		ClOrdID:     clOrdID,
		OrigClOrdID: origClOrdID,
		ExecType:    ExecTypeReplaced,
		Qty:         qty,
		Price:       price,
		Timestamp:   ts,
	}
}

func NewEventID(orderID string, status OrderStatus) string {
	return fmt.Sprintf("%s-%s", orderID, status)
}
