package services

import (
	"errors"
)

var (
	ErrInternal              = errors.New("services: internal error")
	ErrForbidden             = errors.New("services: access forbidden")
	ErrUnauthenticated       = errors.New("services: unauthenticated")
	ErrInvalidCredentials    = errors.New("services: invalid credentials")
	ErrUserNotCreated        = errors.New("services: failed to create user")
	ErrUserAlreadyExists     = errors.New("services: user already exists")
	ErrUserNotFound          = errors.New("services: user was not found")
	ErrItemNotCreated        = errors.New("service: failed to create item")
	ErrItemNotFound          = errors.New("service: item was not found")
	ErrFileNotCreated        = errors.New("service: failed to create file")
	ErrFileNotFound          = errors.New("service: file was not found")
	ErrStopwordNotCreated    = errors.New("service: failed to create stopword")
	ErrStopwordAlreadyExists = errors.New("service: stopword already exists")
	ErrStopwordNotFound      = errors.New("service: stopword was not found")
	ErrTagNotCreated         = errors.New("service: failed to create tag")
	ErrTagNotFound           = errors.New("service: tag was not found")
	ErrTagAlreadyExists      = errors.New("service: tag already exists")
	ErrPdfFileFormat         = errors.New("services: provided file has to be a PDF file")
)

type ServiceError struct {
	Message string
	Kind    error
	Cause   error
}

func (e *ServiceError) Error() string {
	msg := e.Kind.Error()
	if e.Message != "" {
		msg += ": " + e.Message
	}
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}
	return msg
}

func (e *ServiceError) Unwrap() []error {
	return []error{e.Kind, e.Cause}
}

func NewServiceError(kind error, message string, cause error) error {
	return &ServiceError{
		Kind:    kind,
		Message: message,
		Cause:   cause,
	}
}
