package task_events

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type OutboxMessage struct {
	ID       int64
	TaskID   int
	Payload  []byte
	Attempts int
}

type OutboxStore interface {
	ClaimOutboxMessages(ctx context.Context, limit int, lease time.Duration) ([]OutboxMessage, error)
	MarkOutboxPublished(ctx context.Context, id int64) error
	RetryOutboxMessage(ctx context.Context, id int64, availableAt time.Time, cause error) error
}

type EventProducer interface {
	Publish(ctx context.Context, key string, payload []byte) error
}

type OutboxPublisher struct {
	store        OutboxStore
	producer     EventProducer
	logger       *zap.Logger
	batchSize    int
	pollInterval time.Duration
	lease        time.Duration
}

func NewOutboxPublisher(
	store OutboxStore,
	producer EventProducer,
	logger *zap.Logger,
	cfg Config,
) *OutboxPublisher {
	return &OutboxPublisher{
		store:        store,
		producer:     producer,
		logger:       logger,
		batchSize:    cfg.BatchSize,
		pollInterval: cfg.PollInterval,
		lease:        cfg.LeaseDuration,
	}
}

func (p *OutboxPublisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		if err := p.publishBatch(ctx); err != nil && ctx.Err() == nil {
			p.logger.Error("publish task outbox batch", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *OutboxPublisher) publishBatch(ctx context.Context) error {
	messages, err := p.store.ClaimOutboxMessages(ctx, p.batchSize, p.lease)
	if err != nil {
		return fmt.Errorf("claim outbox messages: %w", err)
	}

	for _, message := range messages {
		if err := p.producer.Publish(ctx, fmt.Sprintf("%d", message.TaskID), message.Payload); err != nil {
			retryAt := time.Now().Add(retryDelay(message.Attempts))
			if retryErr := p.store.RetryOutboxMessage(ctx, message.ID, retryAt, err); retryErr != nil {
				p.logger.Error("schedule task event retry", zap.Int64("outbox_id", message.ID), zap.Error(retryErr))
			}
			continue
		}

		if err := p.store.MarkOutboxPublished(ctx, message.ID); err != nil {
			p.logger.Error("mark task event published", zap.Int64("outbox_id", message.ID), zap.Error(err))
		}
	}

	return nil
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 7 {
		attempt = 7
	}
	return time.Second * time.Duration(1<<(attempt-1))
}
