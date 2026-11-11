package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Censtro/todo-api/internal/core/domain"
	core_errors "github.com/Censtro/todo-api/internal/core/errors"
	core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"
	task_events "github.com/Censtro/todo-api/internal/features/tasks/events"
)

func (r *TaskRepository) PatchTask(
	ctx context.Context,
	id int,
	task domain.Task,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Task{}, fmt.Errorf("begin patch task transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	UPDATE todoapp.tasks
	SET 
		title=$1,
		description=$2,
		completed=$3,
		completed_at=$4,
		version = version + 1
	WHERE id=$5 AND version = $6

	RETURNING
		id,
		version,
		title,
		description,
		completed,
		created_at,
		completed_at,
		user_id;
	`

	row := tx.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Completed,
		task.CompletedAt,
		id,
		task.Version,
	)

	var taskModel TaskModel

	err = row.Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.AuthorUserID,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Task{}, fmt.Errorf(
				"Task with id='%d' concurrently accessed: %w",
				id,
				core_errors.ErrConflit,
			)
		}

		return domain.Task{}, fmt.Errorf("scan error: %w", err)
	}

	taskDomain := taskDomainFromModel(taskModel)
	if err := insertTaskOutboxEvent(ctx, tx, task_events.TaskUpdated, taskDomain); err != nil {
		return domain.Task{}, fmt.Errorf("insert task-updated outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Task{}, fmt.Errorf("commit patch task transaction: %w", err)
	}

	return taskDomain, nil
}
