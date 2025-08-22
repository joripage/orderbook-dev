package fixgateway

import (
	"testing"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/fix44/executionreport"
	"github.com/quickfixgo/quickfix"

	"github.com/shopspring/decimal"

	"github.com/joripage/orderbook-dev/pkg/oms/model" // <-- đổi path theo repo của bạn
)

// // ----- Pool setup -----

// var execReportPool = sync.Pool{
// 	New: func() interface{} {
// 		// tạo message base (tránh alloc mới toàn bộ header/body)
// 		return executionreport.New(
// 			field.NewOrderID(""),
// 			field.NewExecID(""),
// 			field.NewExecType(enum.ExecType_NEW),
// 			field.NewOrdStatus(enum.OrdStatus_NEW),
// 			field.NewSide(enum.Side_BUY),
// 			field.NewLeavesQty(decimal.NewFromInt(0), 0),
// 			field.NewCumQty(decimal.NewFromInt(0), 0),
// 			field.NewAvgPx(decimal.NewFromFloat(0), 0),
// 		)
// 	},
// }

// func getExecReport() executionreport.ExecutionReport {
// 	return execReportPool.Get().(executionreport.ExecutionReport)
// }

// func putExecReport(msg executionreport.ExecutionReport) {
// 	resetMessage(msg.Message)
// 	execReportPool.Put(msg)
// }

// func resetMessage(msg *quickfix.Message) {
// 	msg.Header.Init()
// 	msg.Body.Init()
// 	msg.Trailer.Init()
// }

// ----- Bench target function -----

func orderReportToExecutionReportForBenchmark(order *model.Order) quickfix.Messagable {
	execReportMsg := executionreport.New(
		field.NewOrderID(order.OrderID),
		field.NewExecID(order.ExecID),
		field.NewExecType(enum.ExecType(order.ExecType)),
		field.NewOrdStatus(enum.OrdStatus(order.Status)),
		field.NewSide(enum.Side(order.Side)),
		field.NewLeavesQty(decimal.NewFromInt(order.LeavesQuantity), 2),
		field.NewCumQty(decimal.NewFromInt(order.CumQuantity), 2),
		field.NewAvgPx(decimal.NewFromFloat(order.AvgPrice), 2),
	)
	execReportMsg.SetClOrdID(order.GatewayID)
	execReportMsg.SetOrigClOrdID(order.OrigGatewayID)
	execReportMsg.SetAccount(order.Account)
	execReportMsg.SetOrderQty(decimal.NewFromInt(order.Quantity), 0)
	execReportMsg.SetPrice(decimal.NewFromFloat(order.Price), 0)
	execReportMsg.SetTransactTime(order.TransactTime)
	return execReportMsg
}

func orderReportToExecutionReportPool(order *model.Order) quickfix.Messagable {
	execReportMsg := getExecReport()
	execReportMsg.Set(field.NewOrderID(order.OrderID))
	execReportMsg.Set(field.NewExecID(order.ExecID))
	execReportMsg.Set(field.NewExecType(enum.ExecType(order.ExecType)))
	execReportMsg.Set(field.NewOrdStatus(enum.OrdStatus(order.Status)))
	execReportMsg.Set(field.NewSide(enum.Side(order.Side)))
	execReportMsg.Set(field.NewLeavesQty(decimal.NewFromInt(order.LeavesQuantity), 2))
	execReportMsg.Set(field.NewCumQty(decimal.NewFromInt(order.CumQuantity), 2))
	execReportMsg.Set(field.NewAvgPx(decimal.NewFromFloat(order.AvgPrice), 2))
	execReportMsg.SetClOrdID(order.GatewayID)
	execReportMsg.SetOrigClOrdID(order.OrigGatewayID)
	execReportMsg.SetAccount(order.Account)
	execReportMsg.SetOrderQty(decimal.NewFromInt(order.Quantity), 0)
	execReportMsg.SetPrice(decimal.NewFromFloat(order.Price), 0)
	execReportMsg.SetTransactTime(order.TransactTime)

	putExecReport(execReportMsg) // trả về pool
	return execReportMsg
}

// ----- Benchmarks -----

var testOrder = &model.Order{
	OrderID:        "O1",
	ExecID:         "E1",
	ExecType:       "0",
	Status:         "New",
	Side:           "1",
	LeavesQuantity: 100,
	CumQuantity:    0,
	AvgPrice:       100.5,
	GatewayID:      "C1",
	OrigGatewayID:  "C0",
	Account:        "ACC1",
	Quantity:       100,
	Price:          100.5,
	TransactTime:   time.Now(),
}

func BenchmarkExecReportNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = orderReportToExecutionReportForBenchmark(testOrder)
	}
}

func BenchmarkExecReportPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = orderReportToExecutionReportPool(testOrder)
	}
}
