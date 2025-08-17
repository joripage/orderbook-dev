package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joripage/orderbook-dev/pkg/oms"
	fixmanager "github.com/joripage/orderbook-dev/pkg/oms/fix"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// bắt tín hiệu hệ điều hành
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fixGateway := fixmanager.NewFixManager(&fixmanager.FixManagerConfig{
		ConfigFilepath: "./config/fixserver.cfg",
	})
	oms := oms.NewOMS(fixGateway)
	fixGateway.AddOmsInstance(oms)
	oms.Start(ctx)
	fmt.Println("FIX client started. Press Ctrl+C to exit.")

	// chờ signal
	<-sigs
	fmt.Println("Shutting down...")

	// hủy context → các goroutine nhận ctx.Done() sẽ thoát
	cancel()

	fmt.Println("Exited cleanly.")
}
