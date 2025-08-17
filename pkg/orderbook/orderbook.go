// file: pkg/orderbook/orderbook.go

package orderbook

import (
	"container/heap"
	"math"
	"sync"

	"github.com/gammazero/deque"
)

type orderBooker interface {
	addOrder(order *Order)
	registerTradeCallback(fn func(result MatchResult))
}

type orderBookConfig struct {
	EnableLMT     bool // limit
	EnableMTL     bool // market
	EnableIceberg bool // iceberg
	EnableGTC     bool // good till cancel
	EnableIOC     bool // immediate or cancel
	EnableFOK     bool // fill or kill
	// todo
}

type orderBook struct {
	symbol string

	buyOrders  map[float64]*deque.Deque[*Order]
	sellOrders map[float64]*deque.Deque[*Order]

	buyHeap  *PriceHeap
	sellHeap *PriceHeap

	ordersByID map[string]*Order

	icebergMgr icebergHandler

	callbacks []func([]MatchResult)

	mu sync.Mutex
}

type icebergHandler interface {
	addIceberg(*Order)
}

func newOrderBook(symbol string) *orderBook {
	buyHeap := NewPriceHeap(func(i, j float64) bool { return i > j })  // Max-heap
	sellHeap := NewPriceHeap(func(i, j float64) bool { return i < j }) // Min-heap

	ob := &orderBook{
		symbol:     symbol,
		buyOrders:  make(map[float64]*deque.Deque[*Order]),
		sellOrders: make(map[float64]*deque.Deque[*Order]),
		buyHeap:    buyHeap,
		sellHeap:   sellHeap,

		ordersByID: make(map[string]*Order),
	}

	return ob
}

func (ob *orderBook) setIcebergManager(im icebergHandler) {
	ob.icebergMgr = im
}

func (ob *orderBook) addOrder(order *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var results []MatchResult

	switch order.Type {
	case MARKET:
		results = ob.executeMarket(order)
	case LIMIT:
		results = ob.executeLimit(order)
	case ICEBERG:
		results = ob.executeIceberg(order)
	}

	if len(results) > 0 {
		for _, cb := range ob.callbacks {
			cb(results)
		}
	}
}

func (ob *orderBook) cancelOrder(orderID string) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	order, ok := ob.ordersByID[orderID]
	if !ok {
		return false
	}

	var book map[float64]*deque.Deque[*Order]
	var heapRef *PriceHeap
	if order.Side == BUY {
		book = ob.buyOrders
		heapRef = ob.buyHeap
	} else {
		book = ob.sellOrders
		heapRef = ob.sellHeap
	}

	q := book[order.Price]
	if q == nil {
		return false
	}

	for i := 0; i < q.Len(); i++ {
		if q.At(i).ID == orderID {
			q.Remove(i)
			break
		}
	}

	// if price level empty -> delete
	if q.Len() == 0 {
		delete(book, order.Price)
		heapRef.Remove(order.Price)
	}

	delete(ob.ordersByID, orderID)
	return true
}

func (ob *orderBook) modifyOrder(orderID string, newPrice float64, newQty int64) bool {
	ob.mu.Lock()

	order, ok := ob.ordersByID[orderID]
	if !ok {
		return false
	}

	if order.Price == newPrice && newQty < order.Qty {
		order.Qty = newQty
		return true
	}
	ob.mu.Unlock()

	// if quantity increased or price changed -> cancel then add new
	ob.cancelOrder(orderID)
	newOrder := &Order{
		ID:          order.ID,
		Symbol:      order.Symbol,
		Side:        order.Side,
		Price:       newPrice,
		Qty:         newQty,
		Type:        order.Type,
		TimeInForce: order.TimeInForce,
	}
	ob.addOrder(newOrder)

	return true
}

func (ob *orderBook) registerTradeCallback(fn func(result []MatchResult)) {
	ob.callbacks = append(ob.callbacks, fn)
}

func (ob *orderBook) executeMarket(order *Order) []MatchResult {
	order.Price = math.MaxFloat64 // price = MAX for Buy
	if order.Side == SELL {
		order.Price = 0 // price = 0 for Sell
	}

	return ob.executeLimit(order)
}

func (ob *orderBook) executeLimit(order *Order) []MatchResult {
	var results []MatchResult
	var sideBook, counterBook map[float64]*deque.Deque[*Order]
	var sideHeap, counterHeap *PriceHeap
	var priceCompare func(bookPrice, counterPrice float64) bool

	orderQty := order.Qty
	if order.Side == BUY {
		sideBook = ob.buyOrders
		sideHeap = ob.buyHeap
		counterBook = ob.sellOrders
		counterHeap = ob.sellHeap
		priceCompare = func(bookPrice, counterPrice float64) bool { return bookPrice >= counterPrice }
	} else { // SELL
		sideBook = ob.sellOrders
		sideHeap = ob.sellHeap
		counterBook = ob.buyOrders
		counterHeap = ob.buyHeap
		priceCompare = func(bookPrice, counterPrice float64) bool { return bookPrice <= counterPrice }
	}

	results = ob.matchOrder(
		order,
		counterBook,
		counterHeap,
		priceCompare,
		order.Side,
	)

	if order.TimeInForce == IOC {
		return results // don't save remaining qty
	}

	if order.TimeInForce == FOK {
		total := int64(0)
		for _, r := range results {
			total += r.Qty
		}
		if total < orderQty {
			// cancel all
			return nil
		}
		return results
	}

	// GTC add remaining qty to order book
	if order.Qty > 0 {
		ob.addToBook(sideBook, sideHeap, order)
	}

	return results
}

func (ob *orderBook) executeIceberg(order *Order) []MatchResult {
	// send iceberg order to IcebergManager and IcebergManager process itself
	if ob.icebergMgr != nil {
		ob.icebergMgr.addIceberg(order)
	}
	return nil // don't return result immediately
}

func (ob *orderBook) matchOrder(
	order *Order,
	counterBook map[float64]*deque.Deque[*Order],
	counterHeap *PriceHeap,
	priceCompare func(bookPrice, counterPrice float64) bool,
	side Side,
) []MatchResult {
	var results []MatchResult

	for {
		bestPrice, ok := counterHeap.Peek()
		if !ok || !priceCompare(order.Price, bestPrice) {
			break
		}

		q := counterBook[bestPrice]
		if q.Len() == 0 {
			heap.Pop(counterHeap)
			delete(counterBook, bestPrice)
			continue
		}

		best := q.Front()
		q.PopFront()

		matchQty := min(order.Qty, best.Qty)
		order.Qty -= matchQty
		best.Qty -= matchQty

		// bestID come first, then orderID come after that -> orderID = bestID, counterID = orderID, side = side before
		results = append(results, MatchResult{
			OrderID:        best.ID,
			CounterOrderID: order.ID,
			Price:          bestPrice,
			Qty:            matchQty,
			Side: map[Side]Side{
				BUY:  SELL,
				SELL: BUY,
			}[side],
		})

		if best.Qty > 0 {
			q.PushFront(best)
		}

		if order.Qty == 0 {
			return results
		}
	}

	return results
}

func (ob *orderBook) addToBook(book map[float64]*deque.Deque[*Order], priceHeap *PriceHeap, order *Order) {
	if book[order.Price] == nil {
		book[order.Price] = &deque.Deque[*Order]{}
		heap.Push(priceHeap, order.Price)
	}
	book[order.Price].PushBack(order)
	ob.ordersByID[order.ID] = order
}
