package repositories

import (
	"context"
	"qvarkk/kvault/internal/domain"
	"strings"
	"time"
	"unicode"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var itemSortColumns = map[string]string{
	"title":      "i.title",
	"created_at": "i.created_at",
	"updated_at": "i.updated_at",
}

const itemSortDefault = "i.updated_at"

type ItemRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewItemRepo(db *sqlx.DB) *ItemRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &ItemRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *ItemRepo) CreateNew(ctx context.Context, item *domain.Item) error {
	sql, args, err := r.queryBuilder.
		Insert("items").Columns("user_id", "type", "title", "content", "source_url", "url_status").
		Values(item.UserID, item.Type, item.Title, item.Content, item.SourceURL, item.UrlStatus).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(item)
	return toRepositoryError(err)
}

func (r *ItemRepo) List(ctx context.Context, f domain.ListItemFilter) ([]domain.Item, int, error) {
	var items []domain.Item
	var count int

	offset := uint64(f.PageSize * (f.Page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("items i").
		Where(sq.Eq{"i.user_id": f.UserID}).
		Where(sq.Eq{"i.deleted_at": nil})

	var tsQuery string
	if f.Query != "" {
		tsQuery = buildTsQuery(f.Query)
		if tsQuery != "" {
			baseQuery = baseQuery.Where("i.search_vector @@ to_tsquery('simple', ?)", tsQuery)
		}
	}

	var countQuery sq.SelectBuilder
	if len(f.TagIDs) > 0 {
		baseQuery = baseQuery.
			Join("item_tags it ON it.item_id = i.id").
			Where(sq.Eq{"it.tag_id": f.TagIDs})
		countQuery = baseQuery.Column("COUNT(DISTINCT i.id)")
		baseQuery = baseQuery.GroupBy("i.id")
	} else {
		countQuery = baseQuery.Columns("COUNT(*)")
	}

	itemsQuery := baseQuery.Columns("i.*")
	if tsQuery != "" {
		itemsQuery = itemsQuery.OrderByClause(sq.Expr("ts_rank(i.search_vector, to_tsquery('simple', ?)) DESC", tsQuery))
	}
	itemsQuery = itemsQuery.
		OrderBy(safeOrderBy(f.Column, f.Direction, itemSortColumns, itemSortDefault)).
		Offset(offset).
		Limit(uint64(f.PageSize))

	itemsQuerySql, itemsArgs, err := itemsQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	countQuerySql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	if err := selectAndCount(ctx, r.db, itemsQuerySql, itemsArgs, countQuerySql, countArgs, &items, &count); err != nil {
		return nil, 0, err
	}

	return items, count, nil
}

func (r *ItemRepo) GetByID(ctx context.Context, itemID string) (*domain.Item, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("items").
		Where(sq.Eq{"id": itemID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var item domain.Item
	err = r.db.GetContext(ctx, &item, sql, args...)
	return &item, toRepositoryError(err)
}

func (r *ItemRepo) GetActiveByIDForUpdate(ctx context.Context, tx *sqlx.Tx, itemID string) (*domain.Item, error) {
	return r.getByIDForUpdate(ctx, tx, itemID, sq.Eq{"deleted_at": nil})
}

func (r *ItemRepo) GetDeletedByIDForUpdate(ctx context.Context, tx *sqlx.Tx, itemID string) (*domain.Item, error) {
	return r.getByIDForUpdate(ctx, tx, itemID, sq.NotEq{"deleted_at": nil})
}

func (r *ItemRepo) getByIDForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	itemID string,
	deletedCondition sq.Sqlizer,
) (*domain.Item, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("items").
		Where(sq.Eq{"id": itemID}).
		Where(deletedCondition).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var item domain.Item
	err = tx.GetContext(ctx, &item, sql, args...)
	return &item, toRepositoryError(err)
}

func (r *ItemRepo) UpdateTx(ctx context.Context, tx *sqlx.Tx, item *domain.Item) error {
	sql, args, err := r.queryBuilder.
		Update("items").
		Set("title", item.Title).
		Set("content", item.Content).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": item.ID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) SoftDeleteByIDTx(ctx context.Context, tx *sqlx.Tx, itemID string) error {
	sql, args, err := r.queryBuilder.
		Update("items").
		Set("deleted_at", "now()").
		Where(sq.Eq{"id": itemID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) RestoreByIDTx(ctx context.Context, tx *sqlx.Tx, itemID string) error {
	sql, args, err := r.queryBuilder.
		Update("items").
		Set("deleted_at", nil).
		Where(sq.Eq{"id": itemID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) BindTagByItemIDTx(
	ctx context.Context,
	tx *sqlx.Tx,
	itemID, tagID string,
) error {
	sql, args, err := r.queryBuilder.
		Insert("item_tags").
		Columns("item_id", "tag_id", "source").
		Values(itemID, tagID, domain.TagSourceManual).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) UnbindTagByItemIDTx(
	ctx context.Context,
	tx *sqlx.Tx,
	itemID, tagID string,
) error {
	sql, args, err := r.queryBuilder.
		Delete("item_tags").
		Where(sq.Eq{"item_id": itemID}).
		Where(sq.Eq{"tag_id": tagID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) FindIDsByTagID(ctx context.Context, tagID string) ([]string, error) {
	sql, args, err := r.queryBuilder.
		Select("item_id").
		From("item_tags").
		Where(sq.Eq{"tag_id": tagID}).
		ToSql()
	if err != nil {
		return []string{}, toRepositoryError(err)
	}

	var itemIDs []string
	err = r.db.SelectContext(ctx, &itemIDs, sql, args...)
	return itemIDs, toRepositoryError(err)
}

func (r *ItemRepo) ListDeleted(ctx context.Context, f domain.ListItemFilter) ([]domain.Item, int, error) {
	var items []domain.Item
	var count int

	offset := uint64(f.PageSize * (f.Page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("items i").
		Where(sq.Eq{"i.user_id": f.UserID}).
		Where(sq.NotEq{"i.deleted_at": nil})

	countQuery := baseQuery.Columns("COUNT(*)")
	itemsQuery := baseQuery.Columns("i.*").
		OrderBy(safeOrderBy(f.Column, f.Direction, itemSortColumns, itemSortDefault)).
		Offset(offset).
		Limit(uint64(f.PageSize))

	itemsQuerySql, itemsArgs, err := itemsQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	countQuerySql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	if err := selectAndCount(ctx, r.db, itemsQuerySql, itemsArgs, countQuerySql, countArgs, &items, &count); err != nil {
		return nil, 0, err
	}

	return items, count, nil
}

func (r *ItemRepo) PermanentlyDeleteAllDeleted(ctx context.Context, userID string) error {
	sql, args, err := r.queryBuilder.
		Delete("items").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.NotEq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) PermanentlyDeleteByIDTx(ctx context.Context, tx *sqlx.Tx, itemID string) error {
	sql, args, err := r.queryBuilder.
		Delete("items").
		Where(sq.Eq{"id": itemID}).
		Where(sq.NotEq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func buildTsQuery(input string) string {
	tokens := strings.Fields(input)
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		var b strings.Builder
		for _, r := range t {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
				b.WriteRune(r)
			}
		}
		if s := b.String(); s != "" {
			parts = append(parts, s+":*")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " & ")
}
