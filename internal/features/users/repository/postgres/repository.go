package user_postgres_repository

import core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"

type UserRepository struct {
	core_postgres_pool.Pool
}

func NewUserRepository(
	pool core_postgres_pool.Pool,
) *UserRepository {
	return &UserRepository{
		Pool: pool,
	}
}
