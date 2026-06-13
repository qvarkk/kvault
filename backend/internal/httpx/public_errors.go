package httpx

import (
	"errors"
	"net/http"

	"qvarkk/kvault/internal/i18n"

	"github.com/go-playground/validator/v10"
)

var (
	ErrBadRequest          = errors.New("one or more fields have validation errors")
	ErrUnauthorized        = errors.New("wrong credentials")
	ErrForbidden           = errors.New("access to the requested entity is forbidden")
	ErrNotFound            = errors.New("the requested resource was not found")
	ErrUnprocessableEntity = errors.New("the request could not be processed")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInternalServer      = errors.New("an internal server error occurred")
)

var errorStatusMap = map[error]int{
	ErrBadRequest:          http.StatusBadRequest,
	ErrUnauthorized:        http.StatusUnauthorized,
	ErrForbidden:           http.StatusForbidden,
	ErrNotFound:            http.StatusNotFound,
	ErrUnprocessableEntity: http.StatusUnprocessableEntity,
	ErrTooManyRequests:     http.StatusTooManyRequests,
	ErrInternalServer:      http.StatusInternalServerError,
}

type PublicError struct {
	Err              error
	Key              string
	Message          string
	ValidationErrors validator.ValidationErrors
}

type ErrorResponse struct {
	Type       string              `json:"type"`
	Title      string              `json:"title"`
	Status     int                 `json:"status"`
	Instance   string              `json:"instance"`
	Detail     string              `json:"detail"`
	Validation []ValidationDetails `json:"validation,omitempty"`
}

func (e *PublicError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *PublicError) GetHttpStatus() int {
	for err, status := range errorStatusMap {
		if errors.Is(e.Err, err) {
			return status
		}
	}
	return http.StatusInternalServerError
}

func (e *PublicError) ToErrorResponse(instance, locale string) *ErrorResponse {
	status := e.GetHttpStatus()

	var msg string
	if e.Key != "" {
		msg = i18n.Translate(e.Key, locale)
	}
	if msg == "" {
		msg = e.Message
	}
	if msg == "" {
		msg = e.Error()
	}

	return NewErrorResponse(status, instance, msg, e.ValidationErrors, locale)
}

func NewErrorResponse(
	status int,
	instance string,
	detail string,
	validation validator.ValidationErrors,
	locale string,
) *ErrorResponse {
	return &ErrorResponse{
		Type:       "about:blank",
		Title:      http.StatusText(status),
		Status:     status,
		Instance:   instance,
		Detail:     detail,
		Validation: DetailValidationErrors(validation, locale),
	}
}
