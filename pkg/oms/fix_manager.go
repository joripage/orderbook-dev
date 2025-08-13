package oms

import (
	"log"
	"oms-fix/pkg/oms/model"
	"oms-fix/pkg/orderbook"
)

type FixManager struct {
	orderbookManager *orderbook.OrderBookManager
}

func NewFixManager() *FixManager {
	fm := &FixManager{}
	obm := orderbook.NewOrderBookManager(&orderbook.OrderBookManagerConfig{
		EnableIceberg: true,
	})
	cb := func(results []orderbook.MatchResult) {
		// fmt.Println("cb", results)
		for _, r := range results {
			// In vài dòng đầu để kiểm tra
			log.Printf("✅ Match: BUY[%s] <=> SELL[%s] @ %.2f Qty %d\n",
				r.BuyOrderID, r.SellOrderID, r.Price, r.Qty)
		}
		fm.OnOrderReport(results)
	}
	obm.RegisterTradeCallback(cb)

	fm.orderbookManager = obm

	return fm
}

func (s *FixManager) AddOrder(addOrder *model.AddOrder) {
	s.orderbookManager.AddOrder(&orderbook.Order{
		ID:          "1",
		Symbol:      addOrder.Symbol,
		Side:        orderbook.Side(addOrder.Side),
		Price:       addOrder.Price.InexactFloat64(),
		Qty:         int(addOrder.Quantity.IntPart()),
		Type:        orderbook.OrderType(addOrder.Type),
		TimeInForce: orderbook.TimeInForce(addOrder.TimeInForce),
	})
}

func (s *FixManager) ModifyOrder() {

}

func (s *FixManager) CancelOrder() {

}

func (s *FixManager) OnOrderReport(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	if matchResults, ok := args[0].([][]orderbook.MatchResult); ok {
		for m := range matchResults {

		}
	}
}
