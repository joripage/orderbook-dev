package orderbook

import "testing"

func TestCancelOrder(t *testing.T) {
	ob := newOrderBook("test")

	order := &Order{
		ID: "1", Symbol: "test", Side: BUY, Price: 100, Qty: 10, Type: LIMIT,
	}
	ob.addOrder(order)

	if !ob.cancelOrder("1") {
		t.Fatalf("expected cancel success")
	}

	if _, ok := ob.ordersByID["1"]; ok {
		t.Fatalf("order should be removed from ordersByID")
	}
}

func TestModifyOrder_DecreaseQty(t *testing.T) {
	ob := newOrderBook("test")

	order := &Order{
		ID: "1", Symbol: "test", Side: BUY, Price: 100, Qty: 10, Type: LIMIT,
	}
	ob.addOrder(order)

	if !ob.modifyOrder("1", 100, 5) {
		t.Fatalf("expected modify success")
	}

	modified := ob.ordersByID["1"]
	if modified.Qty != 5 {
		t.Fatalf("expected Qty=5, got %d", modified.Qty)
	}
	if modified.Price != 100 {
		t.Fatalf("expected Price=100, got %f", modified.Price)
	}
}

func TestModifyOrder_IncreaseQty(t *testing.T) {
	ob := newOrderBook("test")

	order := &Order{
		ID: "1", Symbol: "test", Side: BUY, Price: 100, Qty: 10, Type: LIMIT,
	}
	ob.addOrder(order)

	if !ob.modifyOrder("1", 100, 20) {
		t.Fatalf("expected modify success")
	}

	modified := ob.ordersByID["1"]
	if modified.Qty != 20 {
		t.Fatalf("expected Qty=20, got %d", modified.Qty)
	}
}

func TestModifyOrder_ChangePrice(t *testing.T) {
	ob := newOrderBook("test")

	order := &Order{
		ID: "1", Symbol: "test", Side: BUY, Price: 100, Qty: 10, Type: LIMIT,
	}
	ob.addOrder(order)

	if !ob.modifyOrder("1", 105, 10) {
		t.Fatalf("expected modify success")
	}

	modified := ob.ordersByID["1"]
	if modified.Price != 105 {
		t.Fatalf("expected Price=105, got %f", modified.Price)
	}
}
