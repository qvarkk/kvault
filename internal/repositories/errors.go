package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"qvarkk/kvault/logger"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	CodeNameUniqueViolation = "unique_violation"
)

var (
	ErrUnknown       = errors.New("repo: unknown database error")
	ErrNotFound      = errors.New("repo: requested entity was not found")
	ErrAlreadyExists = errors.New("repo: entity already exists")
)

func toRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error

	if errors.As(err, &pqErr) {
		switch pqErr.Code.Name() {
		case CodeNameUniqueViolation:
			return wrapError(ErrAlreadyExists, err)
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		return wrapError(ErrNotFound, err)
	}

	logger.Logger.Error("unrecognized DB error was caught", zap.Error(err))
	return wrapError(ErrUnknown, err)
}

func wrapError(sentinelErr error, err error) error {
	return fmt.Errorf("%w: %v", sentinelErr, err)
}
