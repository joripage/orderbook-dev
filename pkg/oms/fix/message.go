package fixgateway

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/quickfixgo/enum"
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

// ----- Pool setup -----

func done(msg *quickfix.Message) {
	// encode / send xong thì trả về pool
	execReportPool.Put(msg)
}

// MessagePool là wrapper để quản lý pool Message
type MessagePool struct {
	pool sync.Pool
}

// NewMessagePool tạo MessagePool mới
func NewMessagePool() *MessagePool {
	return &MessagePool{
		pool: sync.Pool{
			New: func() interface{} {
				// tạo message mới với Header/Body/Trailer init sẵn
				m := quickfix.NewMessage()
				resetMessage(m)
				return m
			},
		},
	}
}

// Get lấy message từ pool (đã reset)
func (mp *MessagePool) Get() *quickfix.Message {
	m := mp.pool.Get().(*quickfix.Message)
	resetMessage(m)
	return m
}

// Put trả message về pool
func (mp *MessagePool) Put(m *quickfix.Message) {
	// reset lại trước khi put để tránh memory leak
	resetMessage(m)
	mp.pool.Put(m)
}

// resetMessage xóa toàn bộ field map và reset tagSort
func resetMessage(m *quickfix.Message) {
	m.Header.Init()
	m.Body.Init()
	m.Trailer.Init()
	m.Header.Clear()
	m.Body.Clear()
	m.Trailer.Clear()
}

var execReportPool = NewMessagePool()

// var newOrderSinglePool = NewMessagePool()

var reportCount = int64(0)

func orderReportToExecutionReport(order model.Order, sessionID *quickfix.SessionID) error {
	atomic.AddInt64(&reportCount, 1)
	fmt.Println(reportCount)

	msg := execReportPool.Get()
	execReportMsg := executionreport.FromMessage(msg)

	// execReportMsg := executionreport.New(
	// 	field.NewOrderID(order.OrderID),
	// 	field.NewExecID(order.ExecID), //think again if it should be in Order model
	// 	field.NewExecType(enum.ExecType(order.ExecType)),
	// 	field.NewOrdStatus(enum.OrdStatus(OrderStatusMapping[order.Status])),
	// 	field.NewSide(enum.Side(SideMapping[order.Side])),
	// 	field.NewLeavesQty(decimal.NewFromInt(order.LeavesQuantity), 2),
	// 	field.NewCumQty(decimal.NewFromInt(order.CumQuantity), 2),
	// 	field.NewAvgPx(decimal.NewFromFloat(order.AvgPrice), 2),
	// )

	execReportMsg.SetMsgType(enum.MsgType_EXECUTION_REPORT)
	execReportMsg.SetOrderID(order.OrderID)
	execReportMsg.SetExecID(order.ExecID) //think again if it should be in Order model
	execReportMsg.SetExecType(enum.ExecType(order.ExecType))
	execReportMsg.SetOrdStatus(enum.OrdStatus(OrderStatusMapping[order.Status]))
	execReportMsg.SetSide(enum.Side(SideMapping[order.Side]))
	execReportMsg.SetLeavesQty(decimal.NewFromInt(order.LeavesQuantity), 2)
	execReportMsg.SetCumQty(decimal.NewFromInt(order.CumQuantity), 2)
	execReportMsg.SetAvgPx(decimal.NewFromFloat(order.AvgPrice), 2)

	execReportMsg.SetClOrdID(order.GatewayID)
	execReportMsg.SetOrigClOrdID(order.OrigGatewayID)
	execReportMsg.SetAccount(order.Account)
	execReportMsg.SetAccountType(enum.AccountType(order.Account))
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
	case model.OrderStatusCanceled:
		execReportMsg.SetExecType(enum.ExecType_CANCELED)
		execReportMsg.SetOrdStatus(enum.OrdStatus_CANCELED)
	case model.OrderStatusReplaced:
		execReportMsg.SetExecType(enum.ExecType_REPLACED)
		execReportMsg.SetOrdStatus(enum.OrdStatus_REPLACED)
	}

	err := quickfix.SendToTarget(execReportMsg, *sessionID)
	if err != nil {
		log.Printf("send err=%v", err)
		return err
	}

	execReportPool.Put(msg)

	return nil
}
