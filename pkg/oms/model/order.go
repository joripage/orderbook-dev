package model

import (
	"time"

	"github.com/shopspring/decimal"
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
	ID int64

	// init info
	Symbol       string
	Side         OrderSide
	Type         OrderType
	TimeInForce  OrderTimeInForce
	Price        decimal.Decimal
	Quantity     decimal.Decimal
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
	CumQuantity    decimal.Decimal
	LeavesQuantity decimal.Decimal
	LastQuantity   decimal.Decimal
	LastPrice      decimal.Decimal
}
