package httpx

import (
	"errors"

	"github.com/lib/pq"
)

const (
	NameUniqueViolation = "unique_violation"
)

// TODO: somehow log unrecognized errors pleeeease
func DBErrorToPublicError(err error) *PublicError {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code.Name() {
		case NameUniqueViolation:
			return &PublicError{
				Err:     ErrUnprocessableEntity,
				Message: parseUniqueConstraintMessage(pqErr.Constraint),
			}
		}
	}

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
