package model

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
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

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
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

var orderEventPool = sync.Pool{
	New: func() interface{} {
		return &OrderEvent{}
	},
}

func NewOrderEvent(order Order, ts time.Time) *OrderEvent {
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

func NewOrderEventUsingPool(order Order, ts time.Time) (*OrderEvent, func()) {
	s := orderEventPool.Get().(*OrderEvent)
	s.EventID = NewEventID(order.OrderID, order.Status)
	s.OrderID = order.OrderID
	s.GatewayID = order.GatewayID
	s.OrigGatewayID = order.OrigGatewayID
	s.OrderStatus = order.Status
	s.ExecType = order.ExecType
	s.Qty = order.Quantity
	s.CumQty = order.CumQuantity
	s.LeavesQty = order.LeavesQuantity
	s.Price = order.Price
	s.ExecID = order.ExecID
	s.LastExecID = order.LastExecID
	s.Timestamp = ts

	resetFn := func() {
		s.EventID = ""
		s.OrderID = ""
		s.GatewayID = ""
		s.OrigGatewayID = ""
		s.OrderStatus = ""
		s.ExecType = ""
		s.Qty = 0
		s.CumQty = 0
		s.LeavesQty = 0
		s.Price = 0
		s.ExecID = ""
		s.LastExecID = ""
		s.Timestamp = time.Time{}
		orderEventPool.Put(s)
	}

	return s, resetFn
}

func NewEventID(orderID string, status OrderStatus) string {
	return fmt.Sprintf("%s-%s", orderID, status)
}
