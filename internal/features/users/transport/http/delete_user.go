package user_transport_http

import (
	"net/http"

	core_logger "github.com/Censtro/todo-api/internal/core/logger"
	core_http_request "github.com/Censtro/todo-api/internal/core/transport/http/request"
	core_http_response "github.com/Censtro/todo-api/internal/core/transport/http/response"
)

// DeleteUser godoc
// @Summary Delete user
// @Tags users
// @Param id path int true "Users ID"
// @Success 204 "User Deleted"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad Request"
// @Failure 404 {object} core_http_response.ErrorResponse "User not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /users/{id} [delete]
func (h *UserHTTPHandler) DeleteUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID path value")
		return
	}

	if err := h.userService.DeleteUser(ctx, userID); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete user")
	}

	responseHandler.NoContentResponse()
}
