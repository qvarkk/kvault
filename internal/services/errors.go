package services

import "errors"

var (
	ErrInternal = errors.New("internal error in service")
	ErrDatabase = errors.New("public error in database")
)

func wrapInternalError(err error) error {
	return errors.Join(ErrInternal, err)
}

func wrapDatabaseError(err error) error {
	return errors.Join(ErrDatabase, err)
}
