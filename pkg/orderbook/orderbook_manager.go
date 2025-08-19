package orderbook

import (
	"sync"
	"time"
)

type OrderBookManagerConfig struct {
	EnableIceberg bool
}

type OrderBookManager struct {
	books     sync.Map
	callbacks []func([]*MatchResult)
	cfg       *OrderBookManagerConfig
}

func NewOrderBookManager(cfg *OrderBookManagerConfig) *OrderBookManager {
	return &OrderBookManager{
		books: sync.Map{},
		cfg:   cfg,
	}
}

func (s *OrderBookManager) AddOrder(order *Order) []*MatchResult {
	book := s.getOrCreateBook(order.Symbol)
	results := book.addOrder(order)
	// if len(results) > 0 {
	// 	for _, cb := range book.callbacks {
	// 		cb(results)
	// 	}
	// }
	return results
}

func (s *OrderBookManager) CancelOrder(symbol, orderID string) error {
	book := s.getOrCreateBook(symbol)
	return book.cancelOrder(orderID)
}

func (s *OrderBookManager) ModifyOrder(symbol, orderID string, newPrice float64, newQty int64) ([]*MatchResult, error) {
	book := s.getOrCreateBook(symbol)
	return book.modifyOrder(orderID, newPrice, newQty)
}

func (s *OrderBookManager) RegisterTradeCallback(cb func([]*MatchResult)) {
	s.callbacks = append(s.callbacks, cb)

	// apply callback to all books
	s.books.Range(func(_, v any) bool {
		book := v.(*orderBook)
		book.registerTradeCallback(cb)
		return true
	})
}

func (s *OrderBookManager) getOrCreateBook(symbol string) *orderBook {
	if val, ok := s.books.Load(symbol); ok {
		return val.(*orderBook)
	}

	book := newOrderBook(symbol)
	for _, cb := range s.callbacks {
		book.registerTradeCallback(cb)
	}

	if s.cfg.EnableIceberg {
		im := newIcebergManager(book, time.Millisecond*1)
		book.setIcebergManager(im)
		im.startScheduler()
	}

	actual, _ := s.books.LoadOrStore(symbol, book)
	return actual.(*orderBook)
}
