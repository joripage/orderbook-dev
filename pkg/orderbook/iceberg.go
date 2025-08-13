package orderbook

import (
	"sync"
	"time"
)

type icebergManager struct {
	mu       sync.Mutex
	book     *orderBook
	orders   map[string]*Order
	interval time.Duration
}

func newIcebergManager(book *orderBook, interval time.Duration) *icebergManager {
	return &icebergManager{
		book:     book,
		orders:   make(map[string]*Order),
		interval: interval,
	}
}

func (im *icebergManager) addIceberg(order *Order) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if order.Type != ICEBERG {
		return
	}

	order.hiddenQty = order.Qty
	im.orders[order.ID] = order

	go im.sliceOnce(order)
}

func (im *icebergManager) sliceOnce(order *Order) {
	if order.hiddenQty <= 0 {
		im.mu.Lock()
		delete(im.orders, order.ID)
		im.mu.Unlock()
		return
	}

	qty := order.VisibleQty
	if qty > order.hiddenQty {
		qty = order.hiddenQty
	}
	order.hiddenQty -= qty

	slice := &Order{
		ID:     order.ID + "-slice-" + time.Now().Format("150405"),
		Symbol: order.Symbol, Side: order.Side, Price: order.Price,
		Qty: qty, Type: LIMIT, TimeInForce: GTC,
	}
	im.book.addOrder(slice)
}

func (im *icebergManager) startScheduler() {
	go func() {
		ticker := time.NewTicker(im.interval)
		for range ticker.C {
			im.mu.Lock()
			for _, order := range im.orders {
				im.sliceOnce(order)
			}
			im.mu.Unlock()
		}
	}()
}
