package main

import (
	"oms-fix/pkg/fixserver"
	"sync"
)

func main() {
	go func() {
		server := fixserver.NewServer()
		// obm := orderbook.NewOrderBookManager(&orderbook.OrderBookManagerConfig{
		// 	EnableIceberg: true,
		// })
		// cb := func(results []orderbook.MatchResult) {
		// 	// fmt.Println("cb", results)
		// 	for _, r := range results {
		// 		// In vài dòng đầu để kiểm tra
		// 		log.Printf("✅ Match: BUY[%s] <=> SELL[%s] @ %.2f Qty %d\n",
		// 			r.BuyOrderID, r.SellOrderID, r.Price, r.Qty)
		// 	}
		// }
		// obm.RegisterTradeCallback(cb)

		// server.SetOrderbookManager(obm)
		server.Init("./config/fixserver.cfg")
		server.Start()
	}()

	// select {}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
