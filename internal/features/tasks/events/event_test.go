package task_events

import (
	"testing"
	"time"

	"github.com/Censtro/todo-api/internal/core/domain"
)

func TestEventEncodeDecodeRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, time.November, 7, 12, 0, 0, 0, time.UTC)
	completedAt := createdAt.Add(time.Hour)
	task := domain.NewTask(42, 3, "Ship Kafka events", nil, true, createdAt, &completedAt, 9)

	encoded, err := New(TaskUpdated, task).Encode()
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if decoded.Type != TaskUpdated || decoded.SchemaVersion != SchemaVersion {
		t.Fatalf("unexpected event metadata: %+v", decoded)
	}
	if decoded.Task.ID != task.ID || decoded.Task.Version != task.Version || decoded.Task.CompletedAt == nil {
		t.Fatalf("unexpected task snapshot: %+v", decoded.Task)
	}
}

func TestEventDecodeRejectsUnsupportedSchema(t *testing.T) {
	event := New(TaskCreated, domain.NewTask(1, 1, "Task", nil, false, time.Now(), nil, 1))
	event.SchemaVersion++
	payload, err := event.Encode()
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}

	if _, err := Decode(payload); err == nil {
		t.Fatal("expected unsupported schema version to fail")
	}
}
