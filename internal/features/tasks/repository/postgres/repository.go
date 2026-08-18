package tasks_postgres_repository

import core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"

type TaskRepository struct {
	pool core_postgres_pool.Pool
}

func NewTaskRepository(
	pool core_postgres_pool.Pool,
) *TaskRepository {
	return &TaskRepository{pool: pool}
}
