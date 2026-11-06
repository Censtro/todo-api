package statistics_service

import (
	"context"

	"github.com/Censtro/todo-api/internal/core/domain"
)

type StatisticsCache interface {
	Get(ctx context.Context, key string) (domain.Statistics, bool, error)
	Set(ctx context.Context, key string, value domain.Statistics) error
}
