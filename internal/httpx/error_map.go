package httpx

import (
	"errors"
	"qvarkk/kvault/internal/services"
)

type mappingRule struct {
	target error
	public *PublicError
}

var serviceErrorRules = []mappingRule{
	{
		target: services.ErrUnauthenticated,
		public: &PublicError{
			Err:     ErrUnauthorized,
			Message: "Invalid or missing API key.",
		},
	},
	{
		target: services.ErrInvalidCredentials,
		public: &PublicError{
			Err:     ErrUnauthorized,
			Message: "Invalid credentials provided.",
		},
	},
	{
		target: services.ErrUserNotFound,
		public: &PublicError{
			Err:     ErrNotFound,
			Message: "User not found.",
		},
	},
	{
		target: services.ErrPdfFileFormat,
		public: &PublicError{
			Err:     ErrUnprocessableEntity,
			Message: "File should be of a PDF content type.",
		},
	},
	{
		target: services.ErrUserAlreadyExists,
		public: &PublicError{
			Err:     ErrUnprocessableEntity,
			Message: "User with this email already exists.",
		},
	},
}

// Does not map errors that cause internal errors.
// Maps only service errors for http layer.
func MapErrorToPublic(err error) *PublicError {
	for _, rule := range serviceErrorRules {
		if errors.Is(err, rule.target) {
			return rule.public
		}
	}
	return &PublicError{
		Err: ErrInternalServer,
	}
}
