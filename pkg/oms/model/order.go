package model

import (
	"time"

	"github.com/joripage/orderbook-dev/pkg/misc"
	"github.com/joripage/orderbook-dev/pkg/oms/constant"
	"github.com/joripage/orderbook-dev/pkg/orderbook"
)

type OrderStatus string

const (
	OrderStatusNew                OrderStatus = "New"
	OrderStatusPartiallyFilled    OrderStatus = "PartiallyFilled"
	OrderStatusFilled             OrderStatus = "Filled"
	OrderStatusDoneForDay         OrderStatus = "DoneForDay"
	OrderStatusCanceled           OrderStatus = "Canceled"
	OrderStatusReplaced           OrderStatus = "Replaced"
	OrderStatusPendingCancel      OrderStatus = "PendingCancel"
	OrderStatusStopped            OrderStatus = "Stopped"
	OrderStatusRejected           OrderStatus = "Rejected"
	OrderStatusSuspended          OrderStatus = "Suspended"
	OrderStatusPendingNew         OrderStatus = "PendingNew"
	OrderStatusCalculated         OrderStatus = "Calculated"
	OrderStatusExpired            OrderStatus = "Expired"
	OrderStatusAcceptedForBidding OrderStatus = "AcceptedForBidding"
	OrderStatusPendingReplace     OrderStatus = "PendingReplace"
)

type OrderExecType string

const (
	ExecTypeNew            OrderExecType = "New"
	ExecTypeDoneForDay     OrderExecType = "DoneForDay"
	ExecTypeCanceled       OrderExecType = "Canceled"
	ExecTypeReplaced       OrderExecType = "Replaced"
	ExecTypePendingCancel  OrderExecType = "PendingCancel"
	ExecTypeStopped        OrderExecType = "Stopped"
	ExecTypeRejected       OrderExecType = "Rejected"
	ExecTypeSuspended      OrderExecType = "Suspended"
	ExecTypePendingNew     OrderExecType = "PendingNew"
	ExecTypeCalculated     OrderExecType = "Calculated"
	ExecTypeExpired        OrderExecType = "Expired"
	ExecTypeRestated       OrderExecType = "Restated"
	ExecTypePendingReplace OrderExecType = "PendingReplace"
	ExecTypeTrade          OrderExecType = "Trade"
	ExecTypeTradeCorrect   OrderExecType = "TradeCorrect"
	ExecTypeTradeCancel    OrderExecType = "TradeCancel"
	ExecTypeOrderStatus    OrderExecType = "OrderStatus"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderType string

const (
	OrderTypeLimit   OrderType = "LIMIT"
	OrderTypeMarket  OrderType = "MARKET"
	OrderTypeIceberg OrderType = "ICEBERG"
)

type OrderTimeInForce string

const (
	OrderTimeInForceDAY OrderTimeInForce = "DAY"
	OrderTimeInForceIOC OrderTimeInForce = "IOC"
	OrderTimeInForceFOK OrderTimeInForce = "FOK"
	OrderTimeInForceGTC OrderTimeInForce = "GTC"
)

type Order struct {
	ID        string
	GatewayID string

	// init info
	Symbol       string
	SecurityID   string
	Exchange     string
	Side         OrderSide
	Type         OrderType
	TimeInForce  OrderTimeInForce
	Price        float64
	Quantity     int64
	Account      string
	TransactTime time.Time

	// counterparty
	CounterpartyAccount string
	CounterpartyExecID  string

	// calculated info
	ExecID         string
	OrderID        string
	Status         OrderStatus
	ExecType       OrderExecType
	CumQuantity    int64
	LeavesQuantity int64
	LastQuantity   int64
	LastPrice      float64
	AvgPrice       float64
}

func (s *Order) UpdateAddOrder(addOrder *AddOrder) {
	qty := addOrder.Quantity.IntPart()
	s.ID = misc.RandSeq(constant.ID_LENGTH)
	s.GatewayID = addOrder.ID
	s.Symbol = addOrder.Symbol
	s.SecurityID = addOrder.SecurityID
	s.Exchange = addOrder.Exchange
	s.Side = addOrder.Side
	s.Type = addOrder.Type
	s.TimeInForce = addOrder.TimeInForce
	s.Price = addOrder.Price.InexactFloat64()
	s.Quantity = qty
	s.Account = addOrder.Account
	s.TransactTime = addOrder.TransactTime

	// calculated info
	s.ExecID = ""
	s.OrderID = s.ID
	s.Status = OrderStatusPendingNew
	s.ExecType = ExecTypeNew
	s.CumQuantity = 0
	s.LeavesQuantity = qty
	s.LastQuantity = 0
	s.LastPrice = 0
	s.AvgPrice = 0
}

func (s *Order) UpdateMatchResult(match *orderbook.MatchResult) {
	oldValue := s.AvgPrice * float64(s.CumQuantity)
	addedValue := match.Price * float64(match.Qty)
	s.LastPrice = match.Price
	s.CumQuantity += match.Qty
	s.LeavesQuantity -= match.Qty
	s.LastQuantity = match.Qty
	s.AvgPrice = (oldValue + addedValue) / float64(s.CumQuantity)
	if s.CumQuantity > 0 {
		s.Status = OrderStatusPartiallyFilled
	}
	if s.LeavesQuantity == 0 {
		s.Status = OrderStatusFilled
	}
	s.ExecID = misc.RandSeq(constant.ID_LENGTH)
}
