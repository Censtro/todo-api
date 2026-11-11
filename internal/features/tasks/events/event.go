package task_events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Censtro/todo-api/internal/core/domain"
	"github.com/google/uuid"
)

const SchemaVersion = 1

type Type string

const (
	TaskCreated Type = "task.created"
	TaskUpdated Type = "task.updated"
	TaskDeleted Type = "task.deleted"
)

type Event struct {
	ID            string       `json:"event_id"`
	Type          Type         `json:"event_type"`
	SchemaVersion int          `json:"schema_version"`
	OccurredAt    time.Time    `json:"occurred_at"`
	Task          TaskSnapshot `json:"task"`
}

type TaskSnapshot struct {
	ID           int        `json:"id"`
	Version      int        `json:"version"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	Completed    bool       `json:"completed"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	AuthorUserID int        `json:"author_user_id"`
}

func New(eventType Type, task domain.Task) Event {
	return Event{
		ID:            uuid.NewString(),
		Type:          eventType,
		SchemaVersion: SchemaVersion,
		OccurredAt:    time.Now().UTC(),
		Task: TaskSnapshot{
			ID:           task.ID,
			Version:      task.Version,
			Title:        task.Title,
			Description:  task.Description,
			Completed:    task.Completed,
			CreatedAt:    task.CreatedAt,
			CompletedAt:  task.CompletedAt,
			AuthorUserID: task.AuthorUserID,
		},
	}
}

func Decode(payload []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return Event{}, fmt.Errorf("decode task event: %w", err)
	}

	if _, err := uuid.Parse(event.ID); err != nil {
		return Event{}, fmt.Errorf("invalid event ID: %w", err)
	}
	if event.SchemaVersion != SchemaVersion {
		return Event{}, fmt.Errorf("unsupported task event schema version: %d", event.SchemaVersion)
	}
	if event.Type != TaskCreated && event.Type != TaskUpdated && event.Type != TaskDeleted {
		return Event{}, fmt.Errorf("unsupported task event type: %q", event.Type)
	}
	if event.Task.ID <= 0 {
		return Event{}, fmt.Errorf("invalid task ID in event: %d", event.Task.ID)
	}

	return event, nil
}

func (event Event) Encode() ([]byte, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("encode task event: %w", err)
	}
	return payload, nil
}
