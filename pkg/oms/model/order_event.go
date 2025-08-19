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
	OrderStatus   OrderStatus
	ExecType      OrderExecType
	Qty           int64
	LeavesQty     int64
	CumQty        int64
	Price         float64
	ExecID        string
	LastExecID    string
	Timestamp     time.Time
}

// func NewOrderEventNewOrder(orderID, gatewayID string, price float64, qty, cumQty, leaveQty int64, ts time.Time) *OrderEvent {
// 	return &OrderEvent{
// 		EventID:   NewEventID(orderID, OrderStatusNew),
// 		OrderID:   orderID,
// 		GatewayID: gatewayID,
// 		ExecType:  ExecTypeNew,
// 		Qty:       qty,
// 		CumQty:    cumQty,
// 		LeaveQty:  leaveQty,
// 		Price:     price,
// 		Timestamp: ts,
// 	}
// }

// func NewOrderEventCancel(orderID, gatewayID, origGatewayID string, ts time.Time) *OrderEvent {
// 	return &OrderEvent{
// 		EventID:       NewEventID(orderID, OrderStatusCanceled),
// 		OrderID:       orderID,
// 		GatewayID:     gatewayID,
// 		OrigGatewayID: origGatewayID,
// 		ExecType:      ExecTypeCanceled,
// 		Timestamp:     ts,
// 	}
// }

// func NewOrderEventCancelReplace(orderID, gatewayID, origGatewayID string, price float64, qty, cumQty, leaveQty int64, ts time.Time) *OrderEvent {
// 	return &OrderEvent{
// 		EventID:       NewEventID(orderID, OrderStatusReplaced),
// 		OrderID:       orderID,
// 		GatewayID:     gatewayID,
// 		OrigGatewayID: origGatewayID,
// 		ExecType:      ExecTypeReplaced,
// 		Qty:           qty,
// 		CumQty:        cumQty,
// 		LeaveQty:      leaveQty,
// 		Price:         price,
// 		Timestamp:     ts,
// 	}
// }

func NewOrderEvent(order *Order, ts time.Time) *OrderEvent {
	return &OrderEvent{
		EventID:       NewEventID(order.OrderID, order.Status),
		OrderID:       order.OrderID,
		GatewayID:     order.GatewayID,
		OrigGatewayID: order.OrigGatewayID,
		OrderStatus:   order.Status,
		ExecType:      order.ExecType,
		Qty:           order.Quantity,
		CumQty:        order.CumQuantity,
		LeavesQty:     order.LeavesQuantity,
		Price:         order.Price,
		ExecID:        order.ExecID,
		LastExecID:    order.LastExecID,
		Timestamp:     ts,
	}
}

func NewEventID(orderID string, status OrderStatus) string {
	return fmt.Sprintf("%s-%s", orderID, status)
}
