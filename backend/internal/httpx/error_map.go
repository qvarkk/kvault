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
			Key:     "err.invalid_api_key",
			Message: "Invalid or missing API key.",
		},
	},
	{
		target: services.ErrForbidden,
		public: &PublicError{
			Err:     ErrForbidden,
			Key:     "err.access_forbidden",
			Message: "Access forbidden.",
		},
	},
	{
		target: services.ErrInvalidCredentials,
		public: &PublicError{
			Err:     ErrUnauthorized,
			Key:     "err.invalid_credentials",
			Message: "Invalid credentials provided.",
		},
	},
	{
		target: services.ErrUserNotFound,
		public: &PublicError{
			Err:     ErrNotFound,
			Key:     "err.user_not_found",
			Message: "User not found.",
		},
	},
	{
		target: services.ErrUserAlreadyExists,
		public: &PublicError{
			Err:     ErrUnprocessableEntity,
			Key:     "err.user_already_exists",
			Message: "User with this username already exists.",
		},
	},
	{
		target: services.ErrItemNotFound,
		public: &PublicError{
			Err:     ErrNotFound,
			Key:     "err.item_not_found",
			Message: "Item with given ID does not exist.",
		},
	},
	{
		target: services.ErrFileNotFound,
		public: &PublicError{
			Err:     ErrNotFound,
			Key:     "err.file_not_found",
			Message: "File with given ID does not exist.",
		},
	},
	{
		target: services.ErrTagNotFound,
		public: &PublicError{
			Err:     ErrNotFound,
			Key:     "err.tag_not_found",
			Message: "This tag does not exist.",
		},
	},
	{
		target: services.ErrTagAlreadyExists,
		public: &PublicError{
			Err:     ErrUnprocessableEntity,
			Key:     "err.tag_already_exists",
			Message: "This tag already exists.",
		},
	},
	{
		target: services.ErrPdfFileFormat,
		public: &PublicError{
			Err:     ErrUnprocessableEntity,
			Key:     "err.pdf_format",
			Message: "File should be of a PDF content type.",
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
