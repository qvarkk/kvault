package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Transactor struct {
	db *sqlx.DB
}

func NewTransactor(db *sqlx.DB) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return toRepositoryError(err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		// since fn is passed from a service this doesn't need toRepositoryError wrapping
		return err
	}

	return toRepositoryError(tx.Commit())
}
