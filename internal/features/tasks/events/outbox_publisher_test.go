package task_events

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

type fakeOutboxStore struct {
	messages  []OutboxMessage
	published []int64
	retries   []retryRecord
}

type retryRecord struct {
	id          int64
	availableAt time.Time
	err         error
}

func (s *fakeOutboxStore) ClaimOutboxMessages(context.Context, int, time.Duration) ([]OutboxMessage, error) {
	return s.messages, nil
}

func (s *fakeOutboxStore) MarkOutboxPublished(_ context.Context, id int64) error {
	s.published = append(s.published, id)
	return nil
}

func (s *fakeOutboxStore) RetryOutboxMessage(_ context.Context, id int64, availableAt time.Time, err error) error {
	s.retries = append(s.retries, retryRecord{id: id, availableAt: availableAt, err: err})
	return nil
}

type fakeProducer struct {
	key     string
	payload []byte
	err     error
}

func (p *fakeProducer) Publish(_ context.Context, key string, payload []byte) error {
	p.key = key
	p.payload = payload
	return p.err
}

func TestOutboxPublisherPublishesAndMarksMessage(t *testing.T) {
	store := &fakeOutboxStore{messages: []OutboxMessage{{ID: 7, TaskID: 42, Payload: []byte(`{"event_id":"id"}`)}}}
	producer := &fakeProducer{}
	publisher := NewOutboxPublisher(store, producer, zap.NewNop(), Config{
		BatchSize: 10, PollInterval: time.Millisecond, LeaseDuration: time.Second,
	})

	if err := publisher.publishBatch(context.Background()); err != nil {
		t.Fatalf("publish batch: %v", err)
	}
	if producer.key != "42" {
		t.Fatalf("expected task ID as Kafka key, got %q", producer.key)
	}
	if len(store.published) != 1 || store.published[0] != 7 {
		t.Fatalf("expected outbox message marked published, got %v", store.published)
	}
	if len(store.retries) != 0 {
		t.Fatalf("did not expect retries, got %v", store.retries)
	}
}

func TestOutboxPublisherSchedulesRetryOnPublishFailure(t *testing.T) {
	publishErr := errors.New("broker unavailable")
	store := &fakeOutboxStore{messages: []OutboxMessage{{ID: 8, TaskID: 43, Attempts: 2, Payload: []byte(`{}`)}}}
	publisher := NewOutboxPublisher(store, &fakeProducer{err: publishErr}, zap.NewNop(), Config{
		BatchSize: 10, PollInterval: time.Millisecond, LeaseDuration: time.Second,
	})
	before := time.Now()

	if err := publisher.publishBatch(context.Background()); err != nil {
		t.Fatalf("publish batch: %v", err)
	}
	if len(store.published) != 0 || len(store.retries) != 1 {
		t.Fatalf("expected failed message to be retried, published=%v retries=%v", store.published, store.retries)
	}
	if store.retries[0].id != 8 || !errors.Is(store.retries[0].err, publishErr) {
		t.Fatalf("unexpected retry record: %+v", store.retries[0])
	}
	if store.retries[0].availableAt.Before(before.Add(1500 * time.Millisecond)) {
		t.Fatalf("retry should use exponential delay, got %s", store.retries[0].availableAt.Sub(before))
	}
}

func TestRetryDelayIsBounded(t *testing.T) {
	if got := retryDelay(1); got != time.Second {
		t.Fatalf("first retry delay = %s, want 1s", got)
	}
	if got := retryDelay(8); got != 64*time.Second {
		t.Fatalf("maximum retry delay = %s, want 64s", got)
	}
}
