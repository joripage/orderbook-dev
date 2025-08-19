package eventstore

import (
	"sync"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type InMemoryEventStore struct {
	mu                     sync.RWMutex
	orders                 map[string][]*model.OrderEvent
	clOrdIDToOrderID       map[string]string // clOrdID -> orderID
	orderIDToLatestClOrdID map[string]string // OrderID -> current ClOrdID
	clOrdToOrigClOrdID     map[string]string // ClOrdID -> OrigClOrdID
	// js      jetstream.JetStream // todo
	// subject string // todo
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		orders:                 make(map[string][]*model.OrderEvent),
		clOrdIDToOrderID:       make(map[string]string),
		orderIDToLatestClOrdID: make(map[string]string),
		clOrdToOrigClOrdID:     make(map[string]string),
		// js:      js,
		// subject: subject,
	}
}

func (s *InMemoryEventStore) AddEvent(ev *model.OrderEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// update store
	s.orders[ev.OrderID] = append(s.orders[ev.OrderID], ev)

	// update ClOrdID chain
	s.TrackClOrdChain(ev.OrderID, ev.ClOrdID, ev.OrigClOrdID)
}

// TrackClOrdChain updates the chain between ClOrdID and OrigClOrdID
func (s *InMemoryEventStore) TrackClOrdChain(orderID, clOrdID, origClOrdID string) {
	// always set the latest ClOrdID
	s.orderIDToLatestClOrdID[orderID] = clOrdID

	// if OrigClOrdID != "" -> create chain
	if origClOrdID != "" {
		s.clOrdToOrigClOrdID[clOrdID] = origClOrdID
	}

	s.clOrdIDToOrderID[clOrdID] = orderID
}

func (s *InMemoryEventStore) GetLatestClOrdID(orderID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.orderIDToLatestClOrdID[orderID]
}

// GetOrigClOrdID returns the immediate OrigClOrdID for a given ClOrdID
func (s *InMemoryEventStore) GetOrigClOrdID(clOrdID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clOrdToOrigClOrdID[clOrdID]
}

func (s *InMemoryEventStore) GetOrderID(clOrdID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clOrdIDToOrderID[clOrdID]
}

// ReconstructChain walks backward to get full chain of ClOrdIDs
func (s *InMemoryEventStore) ReconstructChain(clOrdID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var chain []string
	curr := clOrdID
	for curr != "" {
		chain = append(chain, curr)
		curr = s.clOrdToOrigClOrdID[curr]
	}
	return chain
}
