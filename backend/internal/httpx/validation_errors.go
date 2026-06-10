package httpx

import (
	"qvarkk/kvault/internal/i18n"

	"github.com/go-playground/validator/v10"
)

type ValidationDetails struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func DetailValidationErrors(err validator.ValidationErrors, locale string) []ValidationDetails {
	var details []ValidationDetails
	for _, e := range err {
		details = append(details, ValidationDetails{
			Field:   e.Field(),
			Message: i18n.FormatValidation(e.Tag(), e.Field(), e.Param(), locale),
		})
	}
	return details
}
