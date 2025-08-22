package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/joripage/go_util/pkg/shardqueue"
	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/nats-io/nats.go"
)

type InMemoryEventStore struct {
	mu sync.RWMutex

	gatewayIDToOrderID       map[string]string // gatewayID -> orderID
	orderIDToLatestGatewayID map[string]string // orderID -> current gatewayID
	gatewayIDToOrigGatewayID map[string]string // gatewayID -> origGatewayID

	dispatcher chan *model.OrderEvent

	sq      *shardqueue.Shardqueue
	js      nats.JetStreamContext
	subject string
}

func NewInMemoryEventStore() *InMemoryEventStore {
	nc, _ := nats.Connect(nats.DefaultURL)
	js, _ := nc.JetStream(nats.PublishAsyncMaxPending(65536))
	numShard := 2
	queueSize := 100000
	sq := shardqueue.NewShardQueue(numShard, queueSize)
	store := &InMemoryEventStore{
		gatewayIDToOrderID:       make(map[string]string),
		orderIDToLatestGatewayID: make(map[string]string),
		gatewayIDToOrigGatewayID: make(map[string]string),
		js:                       js,
		subject:                  "ORDERS.*",
		sq:                       sq,
		// dispatcher:               make(chan *model.OrderEvent, queueSize),
	}
	store.sq.Start(func(msg interface{}) error {
		if v, ok := msg.(*model.OrderEvent); ok {
			// _ = v
			store.publish(v)
		}
		return nil
	})
	// go store.runDispatcher()

	return store
}

func (s *InMemoryEventStore) AddEvent(event *model.OrderEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// update store
	// s.orders[event.OrderID] = append(s.orders[event.OrderID], event)

	// update ClOrdID chain
	s.TrackClOrdChain(event.OrderID, event.GatewayID, event.OrigGatewayID)

	s.sq.Shard(event.OrderID, event)
	// s.dispatcher <- event
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

func (s *InMemoryEventStore) publish(event *model.OrderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
		return err
	}
	// ackFuture, err := s.js.PublishAsync("ORDERS.events", data)
	err = publishWithRetry(context.Background(), s.js, "ORDERS.events", data, 5)

	// if err != nil {
	// 	return err
	// }

	// ackFuture.Ok()

	return err
}

func publishWithRetry(ctx context.Context, js nats.JetStreamContext, subj string, msg []byte, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		_, err := js.PublishAsync(subj, msg)
		if err == nil {
			return nil
		}
		backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Printf("Publish failed, retry in %v...", backoff)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("failed to publish after %d retries: %w", maxRetry, err)
}

func (s *InMemoryEventStore) runDispatcher() {
	for msg := range s.dispatcher {
		s.sq.Shard(msg.OrderID, msg)
		// fmt.Println(msg)
		// return
	}

	// for msg := range ch {
	// 	if err := fn(msg); err != nil {
	// 		log.Printf("Shard %d process error: %v", id, err)
	// 	}
	// }
}
