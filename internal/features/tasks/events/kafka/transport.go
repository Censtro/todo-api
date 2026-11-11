package task_events_kafka

import (
	"context"
	"fmt"
	"time"

	task_events "github.com/Censtro/todo-api/internal/features/tasks/events"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, key string, payload []byte) error {
	if err := p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(key), Value: payload}); err != nil {
		return fmt.Errorf("write Kafka task event: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type AuditStore interface {
	RecordTaskEvent(ctx context.Context, event task_events.Event, payload []byte) error
}

type Consumer struct {
	reader *kafka.Reader
	dlq    *kafka.Writer
	store  AuditStore
	logger *zap.Logger
}

func NewConsumer(cfg task_events.Config, store AuditStore, logger *zap.Logger) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			Topic:          cfg.Topic,
			GroupID:        cfg.ConsumerGroup,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 0,
		}),
		dlq: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.DeadLetter,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
		store:  store,
		logger: logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Error("fetch task event from Kafka", zap.Error(err))
			if !wait(ctx, time.Second) {
				return nil
			}
			continue
		}

		if err := c.processUntilStored(ctx, message); err != nil {
			return err
		}
		if err := c.reader.CommitMessages(ctx, message); err != nil {
			c.logger.Error("commit task event offset", zap.Error(err))
		}
	}
}

func (c *Consumer) processUntilStored(ctx context.Context, message kafka.Message) error {
	event, err := task_events.Decode(message.Value)
	if err != nil {
		return c.retryUntilSuccess(ctx, "write invalid task event to dead-letter topic", func() error {
			return c.dlq.WriteMessages(ctx, kafka.Message{
				Key:   message.Key,
				Value: message.Value,
				Headers: []kafka.Header{{
					Key:   "dead-letter-reason",
					Value: []byte(err.Error()),
				}},
			})
		})
	}

	return c.retryUntilSuccess(ctx, "store task audit event", func() error {
		return c.store.RecordTaskEvent(ctx, event, message.Value)
	})
}

func (c *Consumer) retryUntilSuccess(ctx context.Context, operation string, action func() error) error {
	delay := time.Second
	for {
		if err := action(); err == nil {
			return nil
		} else {
			c.logger.Error(operation, zap.Error(err))
		}
		if !wait(ctx, delay) {
			return ctx.Err()
		}
		if delay < 15*time.Second {
			delay *= 2
		}
	}
}

func wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (c *Consumer) Close() error {
	readerErr := c.reader.Close()
	dlqErr := c.dlq.Close()
	if readerErr != nil {
		return fmt.Errorf("close Kafka reader: %w", readerErr)
	}
	if dlqErr != nil {
		return fmt.Errorf("close Kafka dead-letter writer: %w", dlqErr)
	}
	return nil
}
