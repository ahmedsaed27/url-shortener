package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrEmptyBody     = errors.New("request body is required")
	ErrInvalidJSON   = errors.New("invalid JSON")
	ErrMultipleJSON  = errors.New("request body must contain a single JSON object")
	ErrInvalidFields = errors.New("invalid request")
)

func DecodeJSON[T any](r *http.Request, dst *T) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyBody
		}

		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrMultipleJSON
	}

	return nil
}

func Validate(validate *validator.Validate, data any) error {
	if err := validate.Struct(data); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidFields, err)
	}

	return nil
}
