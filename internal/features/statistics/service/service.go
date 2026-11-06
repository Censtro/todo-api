package statistics_service

import (
	"context"
	"time"

	"github.com/Censtro/todo-api/internal/core/domain"
)

type StatisticsService struct {
	statisticsRepository StatisticsRepository
	statisticsCache      StatisticsCache
}

type StatisticsRepository interface {
	GetTasks(
		ctx context.Context,
		userID *int,
		from *time.Time,
		to *time.Time,
	) ([]domain.Task, error)
}

func NewStatisticsService(
	statisticsRepository StatisticsRepository,
	statisticsCache StatisticsCache,
) *StatisticsService {
	return &StatisticsService{
		statisticsRepository: statisticsRepository,
		statisticsCache:      statisticsCache,
	}
}
