package statistics_postgres_repository

import core_postgres_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool"

type StatisticsRepository struct {
	pool core_postgres_pool.Pool
}

func NewStatisticRepository(
	pool core_postgres_pool.Pool,
) *StatisticsRepository {
	return &StatisticsRepository{
		pool: pool,
	}
}
