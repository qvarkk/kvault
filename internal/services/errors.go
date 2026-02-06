package services

import (
	"errors"
	"fmt"
)

var (
	ErrInternal       = errors.New("services: internal error")
	ErrUserNotCreated = errors.New("services: failed to create user")
	ErrUserNotFound   = errors.New("services: user was not found")
)

type ServiceError struct {
	Message string
	Kind    error
	Cause   error
}

func (e *ServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *ServiceError) Unwrap() error {
	return e.Kind
}

func NewServiceError(kind error, message string, cause error) error {
	return &ServiceError{
		Kind:    kind,
		Message: message,
		Cause:   cause,
	}
}
