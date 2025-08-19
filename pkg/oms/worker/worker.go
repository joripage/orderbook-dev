// file: pkg/worker/worker.go
package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/joripage/orderbook-dev/pkg/oms/repo"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type Worker struct {
	order      repo.IOrder
	orderEvent repo.IOrderEvent
}

func NewWorker(repo repo.IRepo) *Worker {
	return &Worker{
		order:      repo.Order(),
		orderEvent: repo.OrderEvent(),
	}
}

func (w *Worker) StartConsumer(ctx context.Context, js nats.JetStreamContext, subject, durable string) error {
	// Create durable consumer
	cons, err := js.PullSubscribe(subject, durable)
	if err != nil {
		return err
	}

	for {
		msgs, err := cons.Fetch(10)
		if err != nil {
			log.Println("Fetch error:", err)
			continue
		}

		for _, msg := range msgs {
			var ev model.OrderEvent
			if err := json.Unmarshal(msg.Data, &ev); err != nil {
				log.Println("unmarshal err", err)
				_ = msg.Ack()
				continue
			}
			if err := w.handleEvent(ev); err != nil {
				log.Println("handleEvent err", err)
				continue
			}
			_ = msg.Ack()
		}
	}
}

func (w *Worker) handleEvent(ev model.OrderEvent) error {
	// insert event
	// _, err := w.db.Exec(`
	// 	INSERT INTO order_events (event_id, order_id, cl_ord_id, exec_type, qty, price, ts)
	// 	VALUES ($1,$2,$3,$4,$5,$6,to_timestamp($7/1000.0))
	// 	ON CONFLICT (event_id) DO NOTHING
	// `, ev.EventID, ev.OrderID, ev.ClOrdID, ev.ExecType, ev.Qty, ev.Price, ev.Timestamp)
	_, err := w.orderEvent.Create(context.Background(), &ev)
	return err
}
