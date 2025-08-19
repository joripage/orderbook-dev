package eventstore

import (
	"sync"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type InMemoryEventStore struct {
	mu                       sync.RWMutex
	orders                   map[string][]*model.OrderEvent
	gatewayIDToOrderID       map[string]string // gatewayID -> orderID
	orderIDToLatestGatewayID map[string]string // orderID -> current gatewayID
	gatewayIDToOrigGatewayID map[string]string // gatewayID -> origGatewayID

	// js      jetstream.JetStream // todo
	// subject string // todo
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		orders:                   make(map[string][]*model.OrderEvent),
		gatewayIDToOrderID:       make(map[string]string),
		orderIDToLatestGatewayID: make(map[string]string),
		gatewayIDToOrigGatewayID: make(map[string]string),
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
	s.TrackClOrdChain(ev.OrderID, ev.GatewayID, ev.OrigGatewayID)
}

// TrackClOrdChain updates the chain between ClOrdID and OrigClOrdID
func (s *InMemoryEventStore) TrackClOrdChain(orderID, gatewayID, origGatewayID string) {
	// always set the latest ClOrdID
	s.orderIDToLatestGatewayID[orderID] = gatewayID

	// if OrigClOrdID != "" -> create chain
	if origGatewayID != "" {
		s.gatewayIDToOrigGatewayID[gatewayID] = origGatewayID
	}

	s.gatewayIDToOrderID[gatewayID] = orderID
}

func (s *InMemoryEventStore) GetLatestGatewayID(orderID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.orderIDToLatestGatewayID[orderID]
}

// GetOrigGatewayID returns the immediate origGatewayID for a given gatewayID
func (s *InMemoryEventStore) GetOrigGatewayID(gatewayID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.gatewayIDToOrigGatewayID[gatewayID]
}

func (s *InMemoryEventStore) GetOrderID(gatewayID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.gatewayIDToOrderID[gatewayID]
}

// ReconstructChain walks backward to get full chain of ClOrdIDs
func (s *InMemoryEventStore) ReconstructChain(gatewayID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var chain []string
	curr := gatewayID
	for curr != "" {
		chain = append(chain, curr)
		curr = s.gatewayIDToOrigGatewayID[curr]
	}
	return chain
}
