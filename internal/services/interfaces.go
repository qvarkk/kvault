package services

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Transactor interface {
	WithTx(context.Context, func(*sqlx.Tx) error) error
}
