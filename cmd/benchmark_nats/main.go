package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)
	js, _ := nc.JetStream(nats.PublishAsyncMaxPending(65536))

	_, _ = js.AddStream(&nats.StreamConfig{
		Name:     "ORDERS",
		Subjects: []string{"ORDERS.*"},
	})

	start := time.Now()
	total := 1_000_000
	for i := range total {
		_ = i
		now := time.Now()
		go func() {
			event := &model.OrderEvent{
				EventID:       "EventID",
				OrderID:       "OrderID",
				GatewayID:     "GatewayID",
				OrigGatewayID: "OrigGatewayID",
				ExecType:      "ExecType",
				OrderStatus:   "OrderStatus",
				Qty:           1000,
				CumQty:        1000,
				LeavesQty:     1000,
				Price:         1000,
				ExecID:        "ExecID",
				LastExecID:    "LastExecID",
				Timestamp:     now,
			}

			data, err := json.Marshal(event)
			if err != nil {
				log.Println("marshal", err)
			}
			ackFuture, err := js.PublishAsync("ORDERS.events", data)
			if err != nil {
				log.Println("publish", err)
			}

			errCh := make(chan error, 100)
			go func(idx int, paf nats.PubAckFuture) {
				select {
				case ack := <-paf.Ok():
					log.Printf("Ack received for msg %d, seq=%d\n", idx, ack.Sequence)
				case err := <-paf.Err():
					log.Printf("Publish failed for msg %d: %v\n", idx, err)
					errCh <- err
				case <-time.After(5 * time.Second):
					log.Printf("Timeout waiting for ack of msg %d\n", idx)
				}
			}(i, ackFuture)

			// ackFuture.Ok()
		}()
	}

	elapsed := time.Since(start)
	msgsPerSec := float64(total) / elapsed.Seconds()

	log.Printf("Sent %d messages in %v", total, elapsed)
	log.Printf("Throughput: %.2f messages/sec", msgsPerSec)
}
