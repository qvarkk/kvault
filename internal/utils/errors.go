package utils

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
)

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

const (
	ErrUniqueViolation = "unique_violation"
)

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

func FormatValidationErrors(err error) []ErrorResponse {
	var errors []ErrorResponse
	for _, e := range err.(validator.ValidationErrors) {
		var element ErrorResponse
		element.FailedField = e.Field()
		element.Tag = e.Tag()
		element.Value = e.Param()
		errors = append(errors, element)
	}

	return errors
}
