package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Order struct {
	ID     int
	Amount int
	Price  int
}

type Trade struct {
	OrderID int
	Amount  int
	Time    int64
}

type LogEntry struct {
	OrderID int
	Text    string
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password by default
		DB:       0,  // Use default DB
	})

	order := Order{ID: 10, Amount: 100, Price: 1000}
	trade := Trade{OrderID: 10, Amount: 10, Time: time.Now().Unix()}
	logEntry := LogEntry{OrderID: 10, Text: "abc"}

	orderKey := fmt.Sprintf("order:%d", order.ID)
	tradeKey := fmt.Sprintf("trade:%d", trade.OrderID)
	logKey := fmt.Sprintf("log:%d", logEntry.OrderID)

	orderJSON, _ := json.Marshal(order)
	tradeJSON, _ := json.Marshal(trade)
	logJSON, _ := json.Marshal(logEntry)

	script := redis.NewScript(`
		if redis.call("EXISTS", KEYS[1]) == 1 then return redis.error_reply("Order key exists") end
		if redis.call("EXISTS", KEYS[2]) == 1 then return redis.error_reply("Trade key exists") end
		if redis.call("EXISTS", KEYS[3]) == 1 then return redis.error_reply("Log key exists") end

		local success = {}
		local ok, err

		ok, err = pcall(function() redis.call("SET", KEYS[1], ARGV[1]) end)
		if not ok then return redis.error_reply("Failed to set order") end
		table.insert(success, KEYS[1])

		ok, err = pcall(function() redis.call("SET", KEYS[2], ARGV[2]) end)
		if not ok then
			for i, k in ipairs(success) do redis.call("DEL", k) end
			return redis.error_reply("Failed to set trade, rollback")
		end
		table.insert(success, KEYS[2])

		ok, err = pcall(function() redis.call("SET", KEYS[3], ARGV[3]) end)
		if not ok then
			for i, k in ipairs(success) do redis.call("DEL", k) end
			return redis.error_reply("Failed to set log, rollback")
		end
		table.insert(success, KEYS[3])

		return "OK"
	`)

	res, err := script.Run(ctx, rdb, []string{orderKey, tradeKey, logKey}, orderJSON, tradeJSON, logJSON).Result()
	if err != nil {
		log.Fatalf("Lua script execution failed: %v", err)
	}

	fmt.Printf("Lua script result: %v\n", res)

	// Concurrent Transaction Benchmark
	const (
		totalOps        = 10
		workers         = 10
		opsPerGoroutine = totalOps / workers
	)

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("tx:key:%d:%d", workerID, i)
				rdb.Watch(ctx, func(tx *redis.Tx) error {
					_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
						pipe.Set(ctx, key, i, 0)
						pipe.Incr(ctx, "tx:counter")
						return nil
					})
					return err
				}, key)
			}
		}(w)
	}

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("Executed %d concurrent transactions in %s (%.2f ops/sec)\n",
		totalOps, duration, float64(totalOps)/duration.Seconds())
}
