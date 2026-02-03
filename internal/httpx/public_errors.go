package httpx

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrBadRequest          = errors.New("One or more fields have validation errors. Please check and try again.")
	ErrUnauthorized        = errors.New("Wrong credentials. Please check and try again.")
	ErrNotFound            = errors.New("The requested resource was not found.")
	ErrUnprocessableEntity = errors.New("The request could not be processed. Please check your input.")
	ErrInternalServer      = errors.New("An internal server error occurred.")
)

var errorStatusMap = map[error]int{
	ErrBadRequest:          http.StatusBadRequest,
	ErrUnauthorized:        http.StatusUnauthorized,
	ErrNotFound:            http.StatusNotFound,
	ErrUnprocessableEntity: http.StatusUnprocessableEntity,
	ErrInternalServer:      http.StatusInternalServerError,
}

type PublicError struct {
	Err              error
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

func (e *PublicError) ToErrorResponse(instance string) *ErrorResponse {
	status := e.GetHttpStatus()

	msg := e.Message
	if msg == "" {
		msg = e.Error()
	}

	return NewErrorResponse(status, instance, msg, e.ValidationErrors)
}

func NewErrorResponse(
	status int,
	instance string,
	detail string,
	validation validator.ValidationErrors,
) *ErrorResponse {
	return &ErrorResponse{
		Type:       "about:blank",
		Title:      http.StatusText(status),
		Status:     status,
		Instance:   instance,
		Detail:     detail,
		Validation: DetailValidationErrors(validation),
	}
}
