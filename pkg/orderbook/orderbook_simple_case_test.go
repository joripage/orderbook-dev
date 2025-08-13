package orderbook

import (
	"fmt"
	"sync"
	"testing"
)

func TestSimpleMatch(t *testing.T) {
	ob := newOrderBook("test")
	cb := func(results []MatchResult) {
		if len(results) != 1 {
			t.Fatalf("expected 1 match, got %d", len(results))
		}

		match := results[0]
		if match.BuyOrderID != "B1" || match.SellOrderID != "S1" {
			t.Errorf("incorrect order IDs in match: %+v", match)
		}
		if match.Qty != 10 || match.Price != 99.0 {
			t.Errorf("incorrect qty/price: %+v", match)
		}
	}
	ob.registerTradeCallback(cb)

	buy := &Order{ID: "B1", Symbol: "ABC", Side: BUY, Price: 100.0, Qty: 10, Type: LIMIT}
	sell := &Order{ID: "S1", Symbol: "ABC", Side: SELL, Price: 99.0, Qty: 10, Type: LIMIT}

	// Add SELL first, then BUY — should match
	ob.addOrder(sell)
	ob.addOrder(buy)

}

func TestNoMatchDueToPrice(t *testing.T) {
	ob := newOrderBook("test")
	cb := func(results []MatchResult) {
		// if there is callback called -> there is trade match -> test failed
		t.Fatalf("expected no match, got %d", len(results))
	}
	ob.registerTradeCallback(cb)

	buy := &Order{ID: "B1", Side: BUY, Price: 98.0, Qty: 10, Type: LIMIT}
	sell := &Order{ID: "S1", Side: SELL, Price: 100.0, Qty: 10, Type: LIMIT}

	ob.addOrder(sell)
	ob.addOrder(buy)
}

func TestPartialMatch(t *testing.T) {
	ob := newOrderBook("test")
	cb := func(results []MatchResult) {
		if len(results) != 1 {
			t.Fatalf("expected 1 match, got %d", len(results))
		}
		if results[0].Qty != 5 {
			t.Errorf("expected matched qty 5, got %d", results[0].Qty)
		}
	}
	ob.registerTradeCallback(cb)

	sell := &Order{ID: "S1", Side: SELL, Price: 100.0, Qty: 5, Type: LIMIT}
	buy := &Order{ID: "B1", Side: BUY, Price: 101.0, Qty: 10, Type: LIMIT}

	ob.addOrder(sell)
	ob.addOrder(buy)
}

func TestFIFOMatch(t *testing.T) {
	ob := newOrderBook("test")
	cb := func(results []MatchResult) {
		if len(results) != 2 {
			t.Fatalf("expected 2 matches, got %d", len(results))
		}
		if results[0].SellOrderID != "S1" || results[1].SellOrderID != "S2" {
			t.Errorf("expected FIFO match order, got %+v", results)
		}
	}
	ob.registerTradeCallback(cb)

	// Add two SELLs at same price
	s1 := &Order{ID: "S1", Side: SELL, Price: 100.0, Qty: 5, Type: LIMIT}
	s2 := &Order{ID: "S2", Side: SELL, Price: 100.0, Qty: 5, Type: LIMIT}
	ob.addOrder(s1)
	ob.addOrder(s2)

	// BUY for total 10, should match in FIFO order: S1 then S2
	buy := &Order{ID: "B1", Side: BUY, Price: 100.0, Qty: 10, Type: LIMIT}
	ob.addOrder(buy)
}

func TestMultiLevelMatch(t *testing.T) {
	ob := newOrderBook("test")
	cb := func(results []MatchResult) {
		if len(results) != 3 {
			t.Fatalf("expected 3 matches, got %d", len(results))
		}
		if results[0].Price != 101.0 || results[2].Price != 103.0 {
			t.Errorf("expected matching from best price, got %+v", results)
		}
	}
	ob.registerTradeCallback(cb)

	// SELLs ở 3 mức giá tăng dần
	sells := []*Order{
		{ID: "S1", Side: SELL, Price: 101.0, Qty: 5, Type: LIMIT},
		{ID: "S2", Side: SELL, Price: 102.0, Qty: 5, Type: LIMIT},
		{ID: "S3", Side: SELL, Price: 103.0, Qty: 5, Type: LIMIT},
	}
	for _, o := range sells {
		ob.addOrder(o)
	}

	// BUY lệnh có giá cao hơn => khớp nhiều mức giá
	buy := &Order{ID: "B1", Side: BUY, Price: 105.0, Qty: 15, Type: LIMIT}
	ob.addOrder(buy)
}

func TestHighVolumeOrders(t *testing.T) {
	ob := newOrderBook("test")
	trade := 0
	cb := func(results []MatchResult) {
		trade += 1
	}
	ob.registerTradeCallback(cb)

	num := 10_000
	for i := 0; i < num; i++ {
		side := BUY
		if i%2 == 0 {
			side = SELL
		}
		order := &Order{
			ID:    fmt.Sprintf("ORD-%d", i),
			Side:  side,
			Price: 100.0,
			Qty:   10,
			Type:  LIMIT,
		}
		ob.addOrder(order)
	}

	if trade != num/2 {
		t.Errorf("expected %d matching, got %d", num/2, trade)
	}
}

func TestConcurrentOrders(t *testing.T) {
	ob := newOrderBook("test")

	var wg sync.WaitGroup
	addOrder := func(id int, side Side) {
		defer wg.Done()
		order := &Order{
			ID:    fmt.Sprintf("C-%d", id),
			Side:  side,
			Price: 100.0,
			Qty:   10,
			Type:  LIMIT,
		}
		ob.addOrder(order)
	}

	n := 1000
	for i := 0; i < n; i++ {
		wg.Add(2)
		go addOrder(i, BUY)
		go addOrder(i, SELL)
	}
	wg.Wait()
	// no crash -> passed
}

func BenchmarkOrderBookMatch(b *testing.B) {
	ob := newOrderBook("test")

	// Pre-load SELL lệnh
	for i := 0; i < 10_000; i++ {
		ob.addOrder(&Order{
			ID:    fmt.Sprintf("SELL-%d", i),
			Side:  SELL,
			Price: 100.0 + float64(i%5),
			Qty:   10,
			Type:  LIMIT,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ob.addOrder(&Order{
			ID:    fmt.Sprintf("BUY-%d", i),
			Side:  BUY,
			Price: 101.0,
			Qty:   10,
			Type:  LIMIT,
		})
	}
}
