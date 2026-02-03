package httpx

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ValidationDetails struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func DetailValidationErrors(err validator.ValidationErrors) []ValidationDetails {
	var details []ValidationDetails
	for _, e := range err {
		var element ValidationDetails
		element.Field = e.Field()
		element.Message = fieldErrorToText(e)
		details = append(details, element)
	}

	return details
}

func fieldErrorToText(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s characters", e.Field(), e.Param())
	case "email":
		return "Invalid email format"
	case "uuid4":
		return fmt.Sprintf("%s must follow uuid4 format", e.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", e.Field(), e.Param())
	}
	return fmt.Sprintf("%s is not valid", e.Field())
}
