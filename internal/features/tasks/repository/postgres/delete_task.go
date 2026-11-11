package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_errors "github.com/Censtro/todo-api/internal/core/errors"
	core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"
	task_events "github.com/Censtro/todo-api/internal/features/tasks/events"
)

func (r *TaskRepository) DeleteTask(
	ctx context.Context,
	id int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete task transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	DELETE FROM todoapp.tasks
	WHERE id = $1
	RETURNING id, version, title, description, completed, created_at, completed_at, user_id;
	`

	var taskModel TaskModel
	if err := tx.QueryRow(ctx, query, id).Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.AuthorUserID,
	); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return fmt.Errorf("task with id='%d': %w", id, core_errors.ErrNotFound)
		}
		return fmt.Errorf("delete task: %w", err)
	}

	task := taskDomainFromModel(taskModel)
	if err := insertTaskOutboxEvent(ctx, tx, task_events.TaskDeleted, task); err != nil {
		return fmt.Errorf("insert task-deleted outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete task transaction: %w", err)
	}

	return nil
}
