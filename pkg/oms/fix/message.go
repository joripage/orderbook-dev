package fixmanager

import (
	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/fix44/executionreport"
	"github.com/quickfixgo/quickfix"
	"github.com/shopspring/decimal"
)

var (
	OrderStatusMapping map[model.OrderStatus]enum.OrdStatus = map[model.OrderStatus]enum.OrdStatus{
		model.OrderStatusNew:                enum.OrdStatus_NEW,
		model.OrderStatusPartiallyFilled:    enum.OrdStatus_PARTIALLY_FILLED,
		model.OrderStatusFilled:             enum.OrdStatus_FILLED,
		model.OrderStatusDoneForDay:         enum.OrdStatus_DONE_FOR_DAY,
		model.OrderStatusCanceled:           enum.OrdStatus_CANCELED,
		model.OrderStatusReplaced:           enum.OrdStatus_REPLACED,
		model.OrderStatusPendingCancel:      enum.OrdStatus_PENDING_CANCEL,
		model.OrderStatusStopped:            enum.OrdStatus_STOPPED,
		model.OrderStatusRejected:           enum.OrdStatus_REJECTED,
		model.OrderStatusSuspended:          enum.OrdStatus_SUSPENDED,
		model.OrderStatusPendingNew:         enum.OrdStatus_PENDING_NEW,
		model.OrderStatusCalculated:         enum.OrdStatus_CALCULATED,
		model.OrderStatusExpired:            enum.OrdStatus_EXPIRED,
		model.OrderStatusAcceptedForBidding: enum.OrdStatus_ACCEPTED_FOR_BIDDING,
		model.OrderStatusPendingReplace:     enum.OrdStatus_PENDING_REPLACE,
	}

	SideMapping map[model.OrderSide]enum.Side = map[model.OrderSide]enum.Side{
		model.OrderSideBuy:  enum.Side_BUY,
		model.OrderSideSell: enum.Side_SELL,
	}
)

// type Order struct {
// 	ID int64

// 	// init info
// 	Symbol       string
// 	Side         OrderSide
// 	Type         OrderType
// 	TimeInForce  OrderTimeInForce
// 	Price        decimal.Decimal
// 	Quantity     decimal.Decimal
// 	Account      string
// 	TransactTime time.Time

// 	// counterparty
// 	CounterpartyAccount string
// 	CounterpartyExecID  string

// 	// calculated info
// 	ExecID         string
// 	OrderID        string
// 	Status         OrderStatus
// 	ExecType       OrderExecType
// 	CumQuantity    decimal.Decimal
// 	LeavesQuantity decimal.Decimal
// 	LastQuantity   decimal.Decimal
// 	LastPrice      decimal.Decimal
// }

func orderReportToExecutionReport(order *model.Order, newOrderSingle *NewOrderSingle) quickfix.Messagable {
	execReportMsg := executionreport.New(
		field.NewOrderID(order.OrderID),
		field.NewExecID(order.ExecID), //think again if it should be in Order model
		field.NewExecType(enum.ExecType(order.ExecType)),
		field.NewOrdStatus(enum.OrdStatus(OrderStatusMapping[order.Status])),
		field.NewSide(enum.Side(SideMapping[order.Side])),
		field.NewLeavesQty(decimal.NewFromInt(order.LeavesQuantity), 2),
		field.NewCumQty(decimal.NewFromInt(order.CumQuantity), 2),
		field.NewAvgPx(decimal.NewFromFloat(order.AvgPrice), 2),
	)

	execReportMsg.SetClOrdID(newOrderSingle.ClOrdID)
	execReportMsg.SetAccount(order.Account)
	execReportMsg.SetAccountType(enum.AccountType(newOrderSingle.Account))
	execReportMsg.SetOrderQty(decimal.NewFromInt(order.Quantity), 0)
	execReportMsg.SetPrice(decimal.NewFromFloat(order.Price), 0)
	execReportMsg.SetTimeInForce(enum.TimeInForce(order.TimeInForce))
	execReportMsg.SetTransactTime(order.TransactTime)
	execReportMsg.SetLastQty(decimal.NewFromInt(order.LastQuantity), 0)
	execReportMsg.SetLastPx(decimal.NewFromFloat(order.LastPrice), 0)
	execReportMsg.SetExecID(order.ExecID)

	switch order.Status {
	case model.OrderStatusPendingNew:
		execReportMsg.SetExecType(enum.ExecType_PENDING_NEW)
		execReportMsg.SetOrdStatus(enum.OrdStatus_PENDING_NEW)
	case model.OrderStatusNew:
		execReportMsg.SetExecType(enum.ExecType_NEW)
		execReportMsg.SetOrdStatus(enum.OrdStatus_NEW)
	case model.OrderStatusPartiallyFilled:
		execReportMsg.SetExecType(enum.ExecType_TRADE)
		execReportMsg.SetOrdStatus(enum.OrdStatus_PARTIALLY_FILLED)
	case model.OrderStatusFilled:
		execReportMsg.SetExecType(enum.ExecType_TRADE)
		execReportMsg.SetOrdStatus(enum.OrdStatus_FILLED)
	}

	if newOrderSingle.MaturityMonthYear != "" {
		execReportMsg.SetMaturityMonthYear(newOrderSingle.MaturityMonthYear)
	}
	if newOrderSingle.SecurityType != "" {
		execReportMsg.SetSecurityType(enum.SecurityType(newOrderSingle.SecurityType))
	}

	execReportMsg.SetTargetCompID(newOrderSingle.SenderCompID)
	execReportMsg.SetSenderCompID(newOrderSingle.TargetCompID)
	if newOrderSingle.OnBehalfOfCompID != "" {
		execReportMsg.SetOnBehalfOfCompID(newOrderSingle.OnBehalfOfCompID)
	}
	if newOrderSingle.DeliverToCompID != "" {
		execReportMsg.SetDeliverToCompID(newOrderSingle.DeliverToCompID)
	}

	return execReportMsg
}
