package response

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type ErrorCode string

const (
	ErrCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeValidation          ErrorCode = "VALIDATION_ERROR"
	ErrCodeInternalServer      ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeUnprocessableEntity ErrorCode = "UNPROCESSABLE_ENTITY"
	ErrCodeTooManyRequests     ErrorCode = "TOO_MANY_REQUESTS"
)

type AppError struct {
	Status  int       `json:"-"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) GetStatus() int {
	return e.Status
}

func NewAppError(status int, code ErrorCode, message string, details any) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}

func BadRequest(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusBadRequest, message, NewAppError(http.StatusBadRequest, ErrCodeBadRequest, message, detail))
}

func Unauthorized(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusUnauthorized, message, NewAppError(http.StatusUnauthorized, ErrCodeUnauthorized, message, detail))
}

func Forbidden(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusForbidden, message, NewAppError(http.StatusForbidden, ErrCodeForbidden, message, detail))
}

func NotFound(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusNotFound, message, NewAppError(http.StatusNotFound, ErrCodeNotFound, message, detail))
}

func Conflict(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusConflict, message, NewAppError(http.StatusConflict, ErrCodeConflict, message, detail))
}

func ValidationError(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusBadRequest, message, NewAppError(http.StatusBadRequest, ErrCodeValidation, message, detail))
}

func InternalServerError(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusInternalServerError, message, NewAppError(http.StatusInternalServerError, ErrCodeInternalServer, message, detail))
}

func ServiceUnavailable(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusServiceUnavailable, message, NewAppError(http.StatusServiceUnavailable, ErrCodeServiceUnavailable, message, detail))
}

func UnprocessableEntity(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusUnprocessableEntity, message, NewAppError(http.StatusUnprocessableEntity, ErrCodeUnprocessableEntity, message, detail))
}

func TooManyRequests(message string, details ...any) error {
	var detail any
	if len(details) > 0 {
		detail = details[0]
	}
	return huma.NewError(http.StatusTooManyRequests, message, NewAppError(http.StatusTooManyRequests, ErrCodeTooManyRequests, message, detail))
}

func FromError(err error) error {
	if err == nil {
		return nil
	}
	return InternalServerError(fmt.Sprintf("An unexpected error occurred: %v", err))
}

func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return InternalServerError(fmt.Sprintf("%s: %v", message, err))
}
