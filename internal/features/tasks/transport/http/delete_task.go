package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Censtro/todo-api/internal/core/logger"
	core_http_request "github.com/Censtro/todo-api/internal/core/transport/http/request"
	core_http_response "github.com/Censtro/todo-api/internal/core/transport/http/response"
)

// DeleteTask godoc
// @Summary Delete task
// @Tags tasks
// @Param id path int true "Task ID"
// @Success 204 "Task Deleted"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad Request"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /tasks/{id} [delete]
func (h *TasksHTTPHandler) DeleteTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	taskID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get TaskID path value",
		)
	}

	if err := h.tasksService.DeleteTask(ctx, taskID); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to delete task",
		)
	}

	responseHandler.NoContentResponse()
}
