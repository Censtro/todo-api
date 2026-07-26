package core_http_request

import (
	"encoding/json"
	"fmt"
	"net/http"

	core_errors "github.com/Censtro/todo-api/internal/core/errors"
	"github.com/go-playground/validator/v10"
)

var requestValidator = validator.New()

type validatable interface {
	Validate() error
}

func DecodeAndValidateRequest(r *http.Request, dto any) error {
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		return fmt.Errorf("decode json: %w", err)

	}

	v, ok := dto.(validatable)

	var err error

	if ok {
		err = v.Validate()
	} else {
		err = requestValidator.Struct(dto)
	}

	if err != nil {
		return fmt.Errorf(
			"request validation: %v: %w",
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}
