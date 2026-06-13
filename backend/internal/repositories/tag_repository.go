package repositories

import (
	"context"
	"time"

	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type ItemTagsByID map[string][]domain.Tag

var tagSortColumns = map[string]string{
	"name":       "t.name",
	"created_at": "t.created_at",
	"updated_at": "t.updated_at",
}

const tagSortDefault = "t.updated_at"

type TagRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewTagRepo(db *sqlx.DB) *TagRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &TagRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *TagRepo) CreateNew(ctx context.Context, tag *domain.Tag) error {
	sql, args, err := r.queryBuilder.
		Insert("tags").
		Columns("user_id", "name").
		Values(tag.UserID, tag.Name).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(tag)
	return toRepositoryError(err)
}

func (r *TagRepo) List(
	ctx context.Context,
	f *domain.ListTagFilter,
) ([]domain.Tag, int, error) {
	page := max(f.Page, 1)
	pageSize := max(f.PageSize, 0)

	//nolint:gosec // G115: page>=1 and pageSize>=0 via max() above
	offset := uint64(pageSize * (page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("tags t").
		LeftJoin("item_tags it ON it.tag_id = t.id").
		LeftJoin("items i ON i.id = it.item_id AND i.deleted_at IS NULL").
		Where(sq.Eq{"t.user_id": f.UserID})

	if f.Query != "" {
		baseQuery = baseQuery.Where(`t.name ILIKE ? ESCAPE '\'`, "%"+escapeLike(f.Query)+"%")
	}

	//nolint:gosec // G115: pageSize>=0 via max() above
	tagsSql, tagsArgs, err := baseQuery.
		Columns("t.*", "COUNT(it.item_id) AS item_count").
		GroupBy("t.id").
		OrderBy(safeOrderBy(f.Column, f.Direction, tagSortColumns, tagSortDefault)).
		Offset(offset).
		Limit(uint64(page)).
		ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	countSql, countArgs, err := baseQuery.Columns("COUNT(DISTINCT t.id)").ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	var tags []domain.Tag
	var count int
	if err := selectAndCount(ctx, r.db, tagsSql, tagsArgs, countSql, countArgs, &tags, &count); err != nil {
		return nil, 0, err
	}

	return tags, count, nil
}

func (r *TagRepo) GetByID(ctx context.Context, tagID string) (*domain.Tag, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("tags").
		Where(sq.Eq{"id": tagID}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var tag domain.Tag
	err = r.db.GetContext(ctx, &tag, sql, args...)
	return &tag, toRepositoryError(err)
}

func (r *TagRepo) GetByIDForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	tagID string,
) (*domain.Tag, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("tags").
		Where(sq.Eq{"id": tagID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var tag domain.Tag
	err = tx.GetContext(ctx, &tag, sql, args...)
	return &tag, toRepositoryError(err)
}

func (r *TagRepo) UpdateTx(ctx context.Context, tx *sqlx.Tx, tag *domain.Tag) error {
	sql, args, err := r.queryBuilder.
		Update("tags").
		Set("name", tag.Name).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": tag.ID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *TagRepo) DeleteByID(ctx context.Context, tagID string) error {
	sql, args, err := r.queryBuilder.
		Delete("tags").
		Where(sq.Eq{"id": tagID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *TagRepo) FindByItemID(ctx context.Context, itemID string) ([]domain.Tag, error) {
	sql, args, err := r.queryBuilder.
		Select("t.*", "it.source").
		From("tags t").
		Join("item_tags it ON it.tag_id = t.id").
		Where(sq.Eq{"it.item_id": itemID}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var tags []domain.Tag
	err = r.db.SelectContext(ctx, &tags, sql, args...)
	return tags, toRepositoryError(err)
}

func (r *TagRepo) FindByItemIDs(
	ctx context.Context,
	itemIDs []string,
) (ItemTagsByID, error) {
	if len(itemIDs) == 0 {
		return ItemTagsByID{}, nil
	}

	sql, args, err := r.queryBuilder.
		Select("t.id", "t.name", "t.user_id", "t.created_at", "t.updated_at", "it.item_id", "it.source").
		From("tags t").
		Join("item_tags it ON it.tag_id = t.id").
		Where(sq.Eq{"it.item_id": itemIDs}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, toRepositoryError(err)
	}
	defer func() { _ = rows.Close() }()

	result := make(ItemTagsByID)
	for rows.Next() {
		var tag domain.Tag
		var itemID string
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.UserID,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&itemID,
			&tag.Source,
		)
		if err != nil {
			return nil, toRepositoryError(err)
		}
		result[itemID] = append(result[itemID], tag)
	}
	if err := rows.Err(); err != nil {
		return nil, toRepositoryError(err)
	}

	return result, nil
}
