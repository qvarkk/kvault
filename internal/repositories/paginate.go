package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

// selectAndCount runs a paged SELECT and its COUNT(*) concurrently, cancelling
// the sibling query (with cause) as soon as either fails. dest must be a pointer
// to a slice; count receives the total. This consolidates the identical
// list+count boilerplate previously copy-pasted across repositories.
func selectAndCount(
	ctx context.Context,
	db *sqlx.DB,
	selectSQL string, selectArgs []any,
	countSQL string, countArgs []any,
	dest any,
	count *int,
) error {
	ctx, cancel := context.WithCancelCause(ctx)
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := db.SelectContext(ctx, dest, selectSQL, selectArgs...); err != nil {
			cancel(err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := db.GetContext(ctx, count, countSQL, countArgs...); err != nil {
			cancel(err)
			return err
		}
		return nil
	})

	_ = g.Wait()

	if cause := context.Cause(ctx); cause != nil {
		return toRepositoryError(cause)
	}
	return nil
}
