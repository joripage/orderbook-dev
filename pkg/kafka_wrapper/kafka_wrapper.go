// kafkakit.go
// A small Go package to publish messages to Kafka and run multiple workers consuming a topic (batch mode version).
//
// In this version, ConsumerGroup.Run() delivers slices of Message (batches) to the handler instead of one by one.

package kafkawrapper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Message struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Time      time.Time
	Headers   map[string]string
	Raw       kafka.Message
}

type ProducerConfig struct {
	Brokers      []string
	Balancer     kafka.Balancer
	BatchSize    int
	BatchBytes   int64
	BatchTimeout time.Duration
	RequiredAcks kafka.RequiredAcks
}

type Producer struct {
	w *kafka.Writer
}

func NewProducer(cfg ProducerConfig) *Producer {
	if cfg.Balancer == nil {
		cfg.Balancer = &kafka.Hash{}
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	if cfg.BatchBytes == 0 {
		cfg.BatchBytes = 1 << 20
	}
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 50 * time.Millisecond
	}
	wr := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               cfg.Balancer,
		BatchSize:              cfg.BatchSize,
		BatchBytes:             cfg.BatchBytes,
		BatchTimeout:           cfg.BatchTimeout,
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireNone,
		Async:                  true,
	}
	return &Producer{w: wr}
}

func (p *Producer) Publish(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	if p == nil || p.w == nil {
		return errors.New("producer not initialized")
	}
	var kh []kafka.Header
	for k, v := range headers {
		kh = append(kh, kafka.Header{Key: k, Value: []byte(v)})
	}
	return p.w.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kh,
		Time:    time.Now(),
	})
}

func (p *Producer) PublishJSON(ctx context.Context, topic string, key string, v any, headers map[string]string) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return p.Publish(ctx, topic, []byte(key), b, headers)
}

func (p *Producer) Close(ctx context.Context) error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}

type ConsumerConfig struct {
	Brokers     []string
	GroupID     string
	Topic       string
	WorkerCount int
	MaxRetries  int
	BackoffMin  time.Duration
	BackoffMax  time.Duration
	DLQTopic    string
	AutoCommit  bool
	// Batch options
	BatchSize    int           // số lượng message tối đa trong 1 batch
	BatchTimeout time.Duration // thời gian tối đa gom batch
}

type ConsumerGroup struct {
	r          *kafka.Reader
	cfg        ConsumerConfig
	prodForDLQ *Producer
}

var ErrSkipCommit = errors.New("skip commit")

func NewConsumerGroup(cfg ConsumerConfig) (*ConsumerGroup, error) {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.BackoffMin == 0 {
		cfg.BackoffMin = 100 * time.Millisecond
	}
	if cfg.BackoffMax == 0 {
		cfg.BackoffMax = 10 * time.Second
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 50
	}
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 200 * time.Millisecond
	}
	if cfg.AutoCommit == false {
	} else {
		cfg.AutoCommit = true
	}

	rd := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		StartOffset: kafka.FirstOffset,
		MaxWait:     500 * time.Millisecond,
		MinBytes:    1,
		MaxBytes:    10 << 20,
	})

	var prod *Producer
	if cfg.DLQTopic != "" {
		prod = NewProducer(ProducerConfig{Brokers: cfg.Brokers})
	}

	return &ConsumerGroup{r: rd, cfg: cfg, prodForDLQ: prod}, nil
}

func (cg *ConsumerGroup) Close() error {
	if cg == nil {
		return nil
	}
	if cg.prodForDLQ != nil {
		_ = cg.prodForDLQ.Close(context.Background())
	}
	if cg.r != nil {
		return cg.r.Close()
	}
	return nil
}

// Run (batch mode): handler receives []Message at a time.
func (cg *ConsumerGroup) Run(ctx context.Context, handler func(context.Context, []Message) error) error {
	if cg == nil || cg.r == nil {
		return errors.New("consumer not initialized")
	}

	batches := make(chan []kafka.Message, cg.cfg.WorkerCount)
	errs := make(chan error, 1)

	// Reader loop to build batches
	go func() {
		defer close(batches)
		var buf []kafka.Message
		timer := time.NewTimer(cg.cfg.BatchTimeout)
		defer timer.Stop()
		for {
			m, err := cg.r.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				errs <- fmt.Errorf("fetch error: %w", err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			buf = append(buf, m)
			if len(buf) >= cg.cfg.BatchSize {
				select {
				case batches <- buf:
					buf = nil
					if !timer.Stop() {
						<-timer.C
					}
					timer.Reset(cg.cfg.BatchTimeout)
				case <-ctx.Done():
					return
				}
			} else {
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(cg.cfg.BatchTimeout)
			}

			select {
			case <-timer.C:
				if len(buf) > 0 {
					batches <- buf
					buf = nil
				}
				timer.Reset(cg.cfg.BatchTimeout)
			default:
			}
		}
	}()

	// Worker pool
	done := make(chan struct{})
	for i := 0; i < cg.cfg.WorkerCount; i++ {
		go func(workerID int) {
			for ms := range batches {
				wrapped := make([]Message, len(ms))
				for i, m := range ms {
					wrapped[i] = wrapMessage(m)
				}
				var attempt int
				for {
					err := handler(ctx, wrapped)
					if err == nil {
						if cg.cfg.AutoCommit {
							_ = cg.r.CommitMessages(ctx, ms...)
						}
						break
					}
					attempt++
					if attempt > cg.cfg.MaxRetries {
						if cg.cfg.DLQTopic != "" && cg.prodForDLQ != nil {
							for _, m := range ms {
								_ = cg.prodForDLQ.Publish(ctx, cg.cfg.DLQTopic, m.Key, m.Value, headersToMap(m.Headers))
							}
						}
						if cg.cfg.AutoCommit {
							_ = cg.r.CommitMessages(ctx, ms...)
						}
						break
					}
					backoff := backoffDuration(cg.cfg.BackoffMin, cg.cfg.BackoffMax, attempt)
					select {
					case <-time.After(backoff):
					case <-ctx.Done():
						return
					}
				}
			}
			done <- struct{}{}
		}(i)
	}

	var workerExited int
	for {
		select {
		case <-done:
			workerExited++
			if workerExited == cg.cfg.WorkerCount {
				return nil
			}
		case err := <-errs:
			_ = err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func wrapMessage(m kafka.Message) Message {
	headers := map[string]string{}
	for _, h := range m.Headers {
		headers[h.Key] = string(h.Value)
	}
	return Message{
		Topic:     m.Topic,
		Partition: m.Partition,
		Offset:    m.Offset,
		Key:       m.Key,
		Value:     m.Value,
		Time:      m.Time,
		Headers:   headers,
		Raw:       m,
	}
}

func headersToMap(hs []kafka.Header) map[string]string {
	out := map[string]string{}
	for _, h := range hs {
		out[h.Key] = string(h.Value)
	}
	return out
}

func backoffDuration(min, max time.Duration, attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	pow := math.Pow(2, float64(attempt-1))
	d := time.Duration(float64(min) * pow)
	if d > max {
		d = max
	}
	if d > 0 {
		d = time.Duration(rand.Int63n(int64(d)))
	}
	return d
}

func HashKey(s string) []byte {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	sum := h.Sum64()
	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[i] = byte(sum >> (56 - 8*i))
	}
	return b
}
