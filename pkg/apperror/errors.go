// Package apperror defines typed application errors with HTTP status codes.
// Use cases return these errors; handlers extract HTTPStatus and Message for the response.
// The underlying Err field is for internal logging only and must not be exposed to clients.
package apperror

import (
	"fmt"
	"net/http"
)

// AppError is an application-level error carrying an HTTP status code,
// a user-facing message, and an optional underlying error for logging.
type AppError struct {
	HTTPStatus int
	Message    string
	Err        error
}

// Error implements the error interface.
// It includes the underlying error when present, for structured log output.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows errors.Is and errors.As to traverse the error chain.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidation returns a 400 Bad Request error.
// Use for missing required fields, invalid formats, or business rule violations on input.
func NewValidation(message string, err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Message:    message,
		Err:        err,
	}
}

// NewNotFound returns a 404 Not Found error.
// resource should be the entity name, e.g. "client", "webhook event".
func NewNotFound(resource string, err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		Err:        err,
	}
}

// NewConflict returns a 409 Conflict error.
// Use for duplicate event_id (idempotency) or unique constraint violations.
func NewConflict(message string, err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusConflict,
		Message:    message,
		Err:        err,
	}
}

// NewInternal returns a 500 Internal Server Error.
// The underlying error is logged internally; the client receives only a generic message.
func NewInternal(err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusInternalServerError,
		Message:    "internal server error",
		Err:        err,
	}
}
