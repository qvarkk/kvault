package errors

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type RFC9457Error struct {
	Type       string            `json:"type"`
	Title      string            `json:"title"`
	Status     int               `json:"status"`
	Instance   string            `json:"instance"`
	Detail     string            `json:"detail"`
	Validation []ValidationError `json:"validation,omitempty"`
}

const (
	ErrUniqueViolation = "unique_violation"
)

func FormRFC9457Error(status int, instance string, detail string) *RFC9457Error {
	err := &RFC9457Error{
		Status:   status,
		Instance: instance,
	}

	switch status {
	case http.StatusBadRequest:
		err.Type = "about:blank"
		err.Title = "Bad Request"
		err.Detail = "One or more fields have validation errors. Please check and try again."
	case http.StatusInternalServerError:
		err.Type = "about:blank"
		err.Title = "Internal Server Error"
		err.Detail = "An internal server error occurred."
	case http.StatusConflict:
		err.Type = "about:blank"
		err.Title = "Conflict"
	case http.StatusUnauthorized:
		err.Type = "about:blank"
		err.Title = "Unauthorized"
		err.Detail = "Wrong credentials. Please check and try again."
	case http.StatusNotFound:
		err.Type = "about:blank"
		err.Title = "Not Found"
	default:
		err.Type = "about:blank"
		err.Title = "Unknown Error"
		err.Detail = "An unknown error occured."
	}

	if detail != "" {
		err.Detail = detail
	}

	return err
}

func ParseDBError(err error) (int, string) {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code.Name() {
		case ErrUniqueViolation:
			return http.StatusConflict, parseUniqueConstraintMessage(pqErr.Constraint)
		}
	}

	return http.StatusInternalServerError, "internal database error"
}

func parseUniqueConstraintMessage(constraint string) string {
	messages := map[string]string{
		"users_email_key": "Email is already in use",
	}
	if msg, found := messages[constraint]; found {
		return msg
	}
	return "Unique constraint error: " + constraint
}

func FormatValidationErrors(err error) []ValidationError {
	var errs []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		var element ValidationError
		element.Field = e.Field()
		element.Message = validationErrorToText(e)
		errs = append(errs, element)
	}

	return errs
}

func validationErrorToText(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s characters", e.Field(), e.Param())
	case "email":
		return "Invalid email format"
	}
	return fmt.Sprintf("%s is not valid", e.Field())
}
