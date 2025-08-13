package orderbook

import (
	"testing"
	"time"
)

func TestLimitOrderMatch(t *testing.T) {
	ob := newOrderBook()
	cb := func(results []MatchResult) {
		if len(results) != 1 || results[0].Qty != 10 {
			t.Errorf("Expected 1 match of 10 units, got %+v", results)
		}
	}
	ob.registerTradeCallback(cb)

	ob.addOrder(&Order{ID: "S1", Side: SELL, Price: 100.0, Qty: 10, Type: LIMIT})
	ob.addOrder(&Order{ID: "B1", Side: BUY, Price: 101.0, Qty: 10, Type: LIMIT})
}

func TestMarketOrderFullMatch(t *testing.T) {
	ob := newOrderBook()
	cb := func(results []MatchResult) {
		if len(results) != 1 || results[0].Qty != 10 {
			t.Errorf("Expected full market match, got %+v", results)
		}
	}
	ob.registerTradeCallback(cb)

	ob.addOrder(&Order{ID: "S1", Side: SELL, Price: 100.0, Qty: 10, Type: LIMIT})
	ob.addOrder(&Order{ID: "B1", Side: BUY, Qty: 10, Type: MARKET})
}

func TestIOCPartialMatch(t *testing.T) {
	ob := newOrderBook()
	cb := func(results []MatchResult) {
		if len(results) != 1 || results[0].Qty != 5 {
			t.Errorf("Expected partial IOC match of 5 units, got %+v", results)
		}
	}
	ob.registerTradeCallback(cb)

	ob.addOrder(&Order{ID: "S1", Side: SELL, Price: 100.0, Qty: 5, Type: LIMIT})
	ob.addOrder(&Order{ID: "B1", Side: BUY, Price: 101.0, Qty: 10, Type: LIMIT, TimeInForce: IOC})
}

func TestFOKRejectPartial(t *testing.T) {
	ob := newOrderBook()
	cb := func(results []MatchResult) {
		if len(results) != 0 {
			t.Errorf("FOK should reject partial fill, got %+v", results)
		}
	}
	ob.registerTradeCallback(cb)

	ob.addOrder(&Order{ID: "S1", Side: SELL, Price: 100.0, Qty: 5, Type: LIMIT})
	ob.addOrder(&Order{ID: "B1", Side: BUY, Price: 101.0, Qty: 10, Type: LIMIT, TimeInForce: FOK})
}

func TestIcebergOrderSlices(t *testing.T) {
	ob := newOrderBook()
	totalMatch := 0
	cb := func(results []MatchResult) {
		for _, result := range results {
			totalMatch += result.Qty
		}
	}
	ob.registerTradeCallback(cb)
	im := newIcebergManager(ob, time.Millisecond*1)
	ob.setIcebergManager(im)
	im.startScheduler()

	ob.addOrder(&Order{
		ID:    "BUY-1",
		Side:  BUY,
		Price: 100.0,
		Qty:   30,
		Type:  LIMIT,
	})

	// Add iceberg order
	ob.addOrder(&Order{
		ID:          "ICE-1",
		Side:        SELL,
		Price:       100.0,
		Qty:         100,
		VisibleQty:  5,
		Type:        ICEBERG,
		TimeInForce: GTC,
	})

	time.Sleep(2 * time.Second) // scheduler running

	if ob.sellOrders[100.0].Len() != 14 { //qty = 100, match 30 => remaining 70 => 70/5 = 14
		t.Errorf("Iceberg expected 14 remaining sell order, got %+v", ob.sellOrders[100.0].Len())
	}
	if totalMatch != 30 {
		t.Errorf("Iceberg expected total match = %d, got %d", 30, totalMatch)
	}
}
