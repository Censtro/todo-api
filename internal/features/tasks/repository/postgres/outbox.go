package tasks_postgres_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Censtro/todo-api/internal/core/domain"
	core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"
	task_events "github.com/Censtro/todo-api/internal/features/tasks/events"
)

func insertTaskOutboxEvent(
	ctx context.Context,
	tx core_postgres_pool.Tx,
	eventType task_events.Type,
	task domain.Task,
) error {
	event := task_events.New(eventType, task)
	payload, err := event.Encode()
	if err != nil {
		return err
	}

	const query = `
	INSERT INTO todoapp.task_outbox (
		event_id, task_id, event_type, schema_version, payload, occurred_at
	) VALUES ($1, $2, $3, $4, $5, $6);`

	_, err = tx.Exec(
		ctx,
		query,
		event.ID,
		event.Task.ID,
		event.Type,
		event.SchemaVersion,
		payload,
		event.OccurredAt,
	)
	return err
}

func (r *TaskRepository) ClaimOutboxMessages(
	ctx context.Context,
	limit int,
	lease time.Duration,
) ([]task_events.OutboxMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin outbox claim transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const selectQuery = `
	SELECT id, task_id, payload, attempts
	FROM todoapp.task_outbox
	WHERE published_at IS NULL
	  AND available_at <= now()
	  AND (lease_until IS NULL OR lease_until < now())
	ORDER BY id
	LIMIT $1
	FOR UPDATE SKIP LOCKED;`

	rows, err := tx.Query(ctx, selectQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("query outbox messages: %w", err)
	}
	defer rows.Close()

	messages := make([]task_events.OutboxMessage, 0, limit)
	for rows.Next() {
		var message task_events.OutboxMessage
		if err := rows.Scan(&message.ID, &message.TaskID, &message.Payload, &message.Attempts); err != nil {
			return nil, fmt.Errorf("scan outbox message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox messages: %w", err)
	}
	rows.Close()

	const leaseQuery = `
	UPDATE todoapp.task_outbox
	SET lease_until = now() + ($2 * interval '1 second'), attempts = attempts + 1
	WHERE id = $1;`
	for i := range messages {
		if _, err := tx.Exec(ctx, leaseQuery, messages[i].ID, lease.Seconds()); err != nil {
			return nil, fmt.Errorf("lease outbox message %d: %w", messages[i].ID, err)
		}
		messages[i].Attempts++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit outbox claim transaction: %w", err)
	}

	return messages, nil
}

func (r *TaskRepository) MarkOutboxPublished(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	const query = `
	UPDATE todoapp.task_outbox
	SET published_at = now(), lease_until = NULL, last_error = NULL
	WHERE id = $1;`
	if _, err := r.pool.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("mark outbox message published: %w", err)
	}
	return nil
}

func (r *TaskRepository) RetryOutboxMessage(ctx context.Context, id int64, availableAt time.Time, cause error) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	const query = `
	UPDATE todoapp.task_outbox
	SET lease_until = NULL, available_at = $2, last_error = $3
	WHERE id = $1;`
	if _, err := r.pool.Exec(ctx, query, id, availableAt, cause.Error()); err != nil {
		return fmt.Errorf("schedule outbox retry: %w", err)
	}
	return nil
}

func (r *TaskRepository) RecordTaskEvent(ctx context.Context, event task_events.Event, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	const query = `
	INSERT INTO todoapp.task_events (
		event_id, task_id, event_type, schema_version, payload, occurred_at
	) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (event_id) DO NOTHING;`
	if _, err := r.pool.Exec(
		ctx,
		query,
		event.ID,
		event.Task.ID,
		event.Type,
		event.SchemaVersion,
		payload,
		event.OccurredAt,
	); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}
