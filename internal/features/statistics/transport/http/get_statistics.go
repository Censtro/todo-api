package statistics_transport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Censtro/todo-api/internal/core/domain"
	core_logger "github.com/Censtro/todo-api/internal/core/logger"
	core_http_request "github.com/Censtro/todo-api/internal/core/transport/http/request"
	core_http_response "github.com/Censtro/todo-api/internal/core/transport/http/response"
)

type GetStatisticsResponse struct {
	TasksCreated               int      `json:"tasks_created"`
	TasksCompleted             int      `json:"tasks_completed"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTime *string  `json:"task_average_completion_time"`
}

func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, from, to, err := getQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get query params",
		)
		return
	}

	statistics, err := h.statisticsService.GetStatistics(ctx, userID, from, to)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get statistics",
		)
		return
	}

	response := toDTOFromDomain(statistics)

	responseHandler.JSONResponse(response, http.StatusOK)
}

func toDTOFromDomain(statistics domain.Statistics) GetStatisticsResponse {
	var avgTime *string
	if statistics.TasksAverageCompletionTime != nil {
		duration := statistics.TasksAverageCompletionTime.String()
		avgTime = &duration
	}

	return GetStatisticsResponse{
		TasksCreated:               statistics.TasksCreated,
		TasksCompleted:             statistics.TasksCompleted,
		TasksCompletedRate:         statistics.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}

func getQueryParams(r *http.Request) (*int, *time.Time, *time.Time, error) {
	userID, err := core_http_request.GetIntQueryParam(r, "user_id")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'userID' query param: %w", err)
	}

	from, err := core_http_request.GetDateQueryParam(r, "from")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'from' quey param: %w", err)
	}

	to, err := core_http_request.GetDateQueryParam(r, "to")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'to' quey param: %w", err)
	}

	return userID, from, to, nil
}
