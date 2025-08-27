// file: pkg/worker/worker.go
package worker

import (
	"context"
	"encoding/json"
	"log"

	kafkawrapper "github.com/joripage/orderbook-dev/pkg/kafka_wrapper"
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
		msgs, err := cons.Fetch(5000)
		if err != nil {
			log.Println("Fetch error:", err)
			continue
		}
		var orderEvents []*model.OrderEvent
		for _, msg := range msgs {
			var orderEvent model.OrderEvent
			if err := json.Unmarshal(msg.Data, &orderEvent); err != nil {
				log.Println("unmarshal err", err)
				_ = msg.Ack()
				continue
			}
			orderEvents = append(orderEvents, &orderEvent)
			_ = msg.Ack()
		}
		if len(orderEvents) > 0 {
			w.orderEvent.BulkCreate(ctx, orderEvents)
		}
	}
}

func (w *Worker) StartConsumerKafka(ctx context.Context, subject, durable string) error {
	// Consumer with 5 workers
	cg, err := kafkawrapper.NewConsumerGroup(kafkawrapper.ConsumerConfig{
		Brokers:     []string{"localhost:29092"},
		GroupID:     "jobs-workers-2",
		Topic:       "ORDERS.events",
		WorkerCount: 5,
		MaxRetries:  5,
		DLQTopic:    "jobs.dlq",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cg.Close()

	log.Println("starting consumers...")
	go func() {
		if err := cg.Run(ctx, func(ctx context.Context, msgs []kafkawrapper.Message) error {
			// // simulate work
			// log.Printf("worker got: key=%s val=%s offset=%d", string(msg.Key), string(msg.Value), msg.Offset)
			// time.Sleep(300 * time.Millisecond)
			// return nil // return an error to trigger retries/DLQ
			var orderEvents []*model.OrderEvent
			for _, msg := range msgs {
				var orderEvent model.OrderEvent
				if err := json.Unmarshal(msg.Value, &orderEvent); err != nil {
					log.Println("unmarshal err", err)
					continue
				}
				orderEvents = append(orderEvents, &orderEvent)
			}
			if len(orderEvents) > 0 {
				w.orderEvent.BulkCreate(ctx, orderEvents)
			}
			return nil
		}); err != nil {
			log.Printf("consumer stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown")

	return nil
}
