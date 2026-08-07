package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "github.com/Censtro/todo-api/internal/core/errors"
)

func GetIntPathValue(r *http.Request, key string) (int, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return 0, fmt.Errorf("no key = '%s in path value: %w", key, core_errors.ErrInvalidArgument)
	}

	val, err := strconv.Atoi(pathValue)
	if err != nil {
		return 0, fmt.Errorf("path value='%s' is not a valid integer: %w", pathValue, core_errors.ErrInvalidArgument)
	}

	return val, nil
}
