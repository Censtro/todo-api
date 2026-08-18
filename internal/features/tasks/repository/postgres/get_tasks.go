package tasks_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Censtro/todo-api/internal/core/domain"
)

func (r *TaskRepository) GetTasks(
	ctx context.Context,
	userID *int,
	limit *int,
	offset *int,
) ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.Timeout())
	defer cancel()

	query := `
	SELECT id, version, title, description, completed, created_at, completed_at, user_id
	FROM todoapp.tasks
	WHERE user_id = COALESCE($1, user_id)
	ORDER BY id ASC
	LIMIT $2
	OFFSET $3`

	rows, err := r.pool.Query(
		ctx,
		query,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var taskModels []TaskModel
	for rows.Next() {
		var TaskModel TaskModel
		err := rows.Scan(
			&TaskModel.ID,
			&TaskModel.Version,
			&TaskModel.Title,
			&TaskModel.Description,
			&TaskModel.Completed,
			&TaskModel.CreatedAt,
			&TaskModel.CompletedAt,
			&TaskModel.AuthorUserID,
		)

		if err != nil {
			return nil, fmt.Errorf("scan tasks: %w", err)
		}

		taskModels = append(taskModels, TaskModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	taskDomains := taskDomainsFromModels(taskModels)
	return taskDomains, nil
}
