package httpx

import (
	"errors"
	"fmt"
	"qvarkk/kvault/logger"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	NameUniqueViolation = "unique_violation"
)

func DBErrorToPublicError(err error) *PublicError {
	var pqErr *pq.Error
	fmt.Printf("%s", err.Error())
	if errors.As(err, &pqErr) {
		switch pqErr.Code.Name() {
		case NameUniqueViolation:
			return &PublicError{
				Err:     ErrUnprocessableEntity,
				Message: parseUniqueConstraintMessage(pqErr.Constraint),
			}
		}
	}

	logger.Logger.Info("unrecognized DB error was caught", zap.Error(err))
	return &PublicError{
		Err: ErrInternalServer,
	}
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
