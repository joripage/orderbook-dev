package main

import (
	"context"
	"encoding/json"
	"flag"

	"github.com/joripage/orderbook-dev/config"
	postgres_wrapper "github.com/joripage/orderbook-dev/pkg/infra/postgres"
	"github.com/joripage/orderbook-dev/pkg/oms/repo"
	"github.com/joripage/orderbook-dev/pkg/oms/worker"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config-file", "", "Specify config file path")
	flag.Parse()

	cfg, err := config.Load(configFile)
	if err != nil {
		panic(err)
	}

	configBytes, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		zap.S().Warnf("could not convert config to JSON: %v", err)
	} else {
		zap.S().Debugf("load config %s", string(configBytes))
	}

	ctx := context.Background()

	// NATS
	nc, _ := nats.Connect(nats.DefaultURL)
	js, _ := nc.JetStream()

	// Ensure stream
	_, _ = js.AddStream(&nats.StreamConfig{
		Name:     "ORDERS",
		Subjects: []string{"ORDERS.*"},
	})

	// InMemoryStore
	// store := oms.NewInMemoryOrderStore(js, "ORDER.events")

	// PostgreSQL
	// db, _ := sql.Open("postgres", "postgres://user:pass@localhost:5432/oms?sslmode=disable")

	// init db
	db, err := postgres_wrapper.InitPostgres(cfg.OmsDB)
	if err != nil {
		zap.S().Errorf("init db fail with err: %v", err)
		panic(err)
	}

	// init repo
	sqlRepo := repo.NewRepo(db)
	if err != nil {
		zap.S().Errorf("Init repo error: %v", err)
		panic(err)
	}

	// Worker
	w := worker.NewWorker(sqlRepo)
	go w.StartConsumer(ctx, js, "ORDERS.events", "order_worker")

	// Add test order
	// store.AddOrder(ctx, &oms.Order{
	// 	OrderID: "O1", ClOrdID: "C1", Symbol: "AAPL", Side: oms.BUY, Price: 150.0, Qty: 100,
	// })

	select {}
}
