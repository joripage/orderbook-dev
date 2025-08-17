package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/joripage/orderbook-dev/pkg/orderbook"
)

const (
	numOrders = 1_000_000
	minPrice  = 100.0
	maxPrice  = 200.0
	minQty    = 1
	maxQty    = 100
)

func randomOrder(id int) *orderbook.Order {
	side := orderbook.BUY
	if rand.Intn(2) == 0 {
		side = orderbook.SELL
	}
	price := minPrice + rand.Float64()*(maxPrice-minPrice)
	qty := int64(rand.Intn(maxQty-minQty+1) + minQty)

	return &orderbook.Order{
		ID:     fmt.Sprintf("ORD-%06d", id),
		Symbol: "ABC",
		Side:   side,
		Price:  float64(int(price*100)) / 100, // làm tròn 2 chữ số thập phân
		Qty:    qty,
		Type:   orderbook.LIMIT,
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	obm := orderbook.NewOrderBookManager(&orderbook.OrderBookManagerConfig{
		EnableIceberg: true,
	})
	totalMatched := 0
	totalQty := int64(0)
	cb := func(results []orderbook.MatchResult) {
		// fmt.Println("cb", results)
		for _, r := range results {
			totalMatched++
			totalQty += r.Qty
			// In vài dòng đầu để kiểm tra
			if totalMatched <= 5 {
				log.Printf("✅ Match: BUY[%s] <=> SELL[%s] @ %.2f Qty %d\n",
					r.OrderID, r.CounterOrderID, r.Price, r.Qty)
			}
		}
	}
	obm.RegisterTradeCallback(cb)

	start := time.Now()
	for i := 0; i < numOrders; i++ {
		order := randomOrder(i + 1)
		obm.AddOrder(order)
	}

	elapsed := time.Since(start)

	fmt.Println("--------")
	fmt.Printf("🏁 Total Orders     : %d\n", numOrders)
	fmt.Printf("✅ Total Matches    : %d\n", totalMatched)
	fmt.Printf("📦 Total Matched Qty: %d\n", totalQty)
	fmt.Printf("⏱️ Time Taken       : %s\n", elapsed)
}
