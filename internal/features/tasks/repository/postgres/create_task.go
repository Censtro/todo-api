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

func (r *TaskRepository) CreateTask(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Task{}, fmt.Errorf("begin create task transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	INSERT INTO todoapp.tasks (title, description, completed, created_at, completed_at, user_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, version, title, description, completed, created_at, completed_at, user_id;`

	row := tx.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Completed,
		task.CreatedAt,
		task.CompletedAt,
		task.AuthorUserID,
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
		if errors.Is(err, core_postgres_pool.ErrViolatesForeignKey) {
			return domain.Task{}, fmt.Errorf(
				"%v: user with id='%d': %w",
				err,
				task.AuthorUserID,
				core_errors.ErrNotFound,
			)
		}

		return domain.Task{}, fmt.Errorf("scan error %w", err)
	}

	taskDomain := taskDomainFromModel(taskModel)
	if err := insertTaskOutboxEvent(ctx, tx, task_events.TaskCreated, taskDomain); err != nil {
		return domain.Task{}, fmt.Errorf("insert task-created outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Task{}, fmt.Errorf("commit create task transaction: %w", err)
	}

	return taskDomain, nil
}
