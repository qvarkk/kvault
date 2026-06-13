package services

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type ItemRepo interface {
	CreateNew(context.Context, *domain.Item) error
	List(context.Context, domain.ListItemFilter) ([]domain.Item, int, error)
	GetByID(context.Context, string) (*domain.Item, error)
	GetActiveByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.Item, error)
	GetDeletedByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.Item, error)
	UpdateTx(context.Context, *sqlx.Tx, *domain.Item) error
	SoftDeleteByIDTx(context.Context, *sqlx.Tx, string) error
	RestoreByIDTx(context.Context, *sqlx.Tx, string) error
	BindTagByItemIDTx(ctx context.Context, tx *sqlx.Tx, itemID, tagID string) error
	UnbindTagByItemIDTx(ctx context.Context, tx *sqlx.Tx, itemID, tagID string) error
	FindIDsByTagID(context.Context, string) ([]string, error)
	ListDeleted(context.Context, domain.ListItemFilter) ([]domain.Item, int, error)
	PermanentlyDeleteAllDeleted(ctx context.Context, userID string) error
	PermanentlyDeleteByIDTx(context.Context, *sqlx.Tx, string) error
}

type ItemService struct {
	itemRepo     ItemRepo
	tagRepo      TagRepo
	transactor   Transactor
	cache        CacheStore
	enqueuer     TaskEnqueuer
	itemCacheTtl time.Duration
}

type CreateItemInput struct {
	UserID    string
	Title     string
	Content   string
	SourceURL string
}

type UpdateItemInput struct {
	ItemID  string
	UserID  string
	Title   *string
	Content *string
}

func NewItemService(
	itemRepo ItemRepo,
	tagRepo TagRepo,
	transactor Transactor,
	cache CacheStore,
	enqueuer TaskEnqueuer,
	cacheTtl time.Duration,
) *ItemService {
	return &ItemService{
		itemRepo:     itemRepo,
		tagRepo:      tagRepo,
		transactor:   transactor,
		cache:        cache,
		enqueuer:     enqueuer,
		itemCacheTtl: cacheTtl,
	}
}

func (s *ItemService) CreateNew(ctx context.Context, input CreateItemInput) (*domain.Item, error) {
	item := &domain.Item{
		UserID:    input.UserID,
		Type:      domain.ItemTypeText,
		Title:     input.Title,
		Content:   NewNullString(input.Content),
		SourceURL: NewNullString(input.SourceURL),
	}

	err := s.itemRepo.CreateNew(ctx, item)
	if err != nil {
		return nil, NewServiceError(ErrItemNotCreated, "database error", err)
	}

	invalidateListCache(ctx, s.cache, itemListVersionKey(input.UserID))

	return item, nil
}

func (s *ItemService) List(ctx context.Context, f domain.ListItemFilter) ([]domain.Item, int, error) {
	version := listVersion(ctx, s.cache, itemListVersionKey(f.UserID))
	cacheKey := itemListKey(version, f)

	if cached := getFromCache[cachedList[domain.Item]](ctx, s.cache, cacheKey); cached != nil {
		return cached.Entities, cached.Count, nil
	}

	items, count, err := s.itemRepo.List(ctx, f)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list items internal error", err)
	}

	if len(items) > 0 {
		ids := make([]string, len(items))
		for i, item := range items {
			ids[i] = item.ID
		}

		tagsByItem, err := s.tagRepo.FindByItemIDs(ctx, ids)
		if err != nil {
			return nil, 0, NewServiceError(ErrInternal, "get item tags internal error", err)
		}

		for i := range items {
			items[i].Tags = tagsByItem[items[i].ID]
		}
	}

	setToCache(ctx, s.cache, cacheKey, cachedList[domain.Item]{Entities: items, Count: count}, s.itemCacheTtl)

	return items, count, nil
}

func (s *ItemService) GetByID(ctx context.Context, itemID, userID string) (*domain.Item, error) {
	cacheKey := itemKey(itemID)
	if item := getFromCache[domain.Item](ctx, s.cache, cacheKey); item != nil {
		if item.UserID != userID {
			return nil, NewServiceError(ErrItemNotFound, "forbidden", nil)
		}
		return item, nil
	}

	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, NewServiceError(ErrItemNotFound, "not found", err)
	}

	if item.UserID != userID {
		return nil, NewServiceError(ErrItemNotFound, "forbidden", nil)
	}

	tags, err := s.tagRepo.FindByItemID(ctx, itemID)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "get item tags internal error", err)
	}
	item.Tags = append(item.Tags, tags...)

	setToCache(ctx, s.cache, cacheKey, item, s.itemCacheTtl)

	return item, nil
}

func (s *ItemService) Update(ctx context.Context, input UpdateItemInput) (*domain.Item, error) {
	var updated *domain.Item

	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		item, err := s.itemRepo.GetActiveByIDForUpdate(ctx, tx, input.ItemID)
		if err != nil {
			return NewServiceError(ErrItemNotFound, "not found", err)
		}

		if item.UserID != input.UserID {
			return NewServiceError(ErrItemNotFound, "forbidden", nil)
		}

		if input.Title != nil {
			item.Title = *input.Title
		}
		if input.Content != nil {
			item.Content = NewNullString(*input.Content)
		}

		if err := s.itemRepo.UpdateTx(ctx, tx, item); err != nil {
			return NewServiceError(ErrInternal, "update item internal error", err)
		}

		updated = item
		return nil
	})
	if err != nil {
		return nil, err
	}

	invalidateSingleCache(ctx, s.cache, itemKey(input.ItemID))
	invalidateListCache(ctx, s.cache, itemListVersionKey(input.UserID))

	tags, err := s.tagRepo.FindByItemID(ctx, updated.ID)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "get item tags internal error", err)
	}
	updated.Tags = tags

	return updated, nil
}

func (s *ItemService) DeleteByID(ctx context.Context, itemID, userID string) error {
	return s.authorizeAndMutateTx(
		ctx, itemID, userID,
		s.itemRepo.GetActiveByIDForUpdate,
		s.itemRepo.SoftDeleteByIDTx,
	)
}

func (s *ItemService) RestoreByID(ctx context.Context, itemID, userID string) error {
	return s.authorizeAndMutateTx(
		ctx, itemID, userID,
		s.itemRepo.GetDeletedByIDForUpdate,
		s.itemRepo.RestoreByIDTx,
	)
}

func (s *ItemService) authorizeAndMutateTx(
	ctx context.Context, itemID, userID string,
	getFn func(context.Context, *sqlx.Tx, string) (*domain.Item, error),
	mutateFn func(context.Context, *sqlx.Tx, string) error,
) error {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		item, err := getFn(ctx, tx, itemID)
		if err != nil {
			return NewServiceError(ErrItemNotFound, "not found", err)
		}

		if item.UserID != userID {
			return NewServiceError(ErrItemNotFound, "forbidden", nil)
		}

		err = mutateFn(ctx, tx, itemID)
		if err != nil {
			return NewServiceError(ErrInternal, "mutate item internal error", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	invalidateSingleCache(ctx, s.cache, itemKey(itemID))
	invalidateListCache(ctx, s.cache, itemListVersionKey(userID))

	return nil
}

func (s *ItemService) AttachTagByItemID(
	ctx context.Context,
	itemID, tagID, userID string,
) error {
	return s.authorizeAndBindTagTx(ctx, itemID, tagID, userID, s.itemRepo.BindTagByItemIDTx)
}

func (s *ItemService) DetachTagByItemID(
	ctx context.Context,
	itemID, tagID, userID string,
) error {
	return s.authorizeAndBindTagTx(ctx, itemID, tagID, userID, s.itemRepo.UnbindTagByItemIDTx)
}

func (s *ItemService) authorizeAndBindTagTx(
	ctx context.Context,
	itemID, tagID, userID string,
	bindFn func(ctx context.Context, tx *sqlx.Tx, itemID, tagID string) error,
) error {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		item, err := s.itemRepo.GetActiveByIDForUpdate(ctx, tx, itemID)
		if err != nil {
			return NewServiceError(ErrItemNotFound, "not found", err)
		}

		tag, err := s.tagRepo.GetByIDForUpdate(ctx, tx, tagID)
		if err != nil {
			return NewServiceError(ErrTagNotFound, "not found", err)
		}

		if item.UserID != tag.UserID || item.UserID != userID {
			return NewServiceError(ErrItemNotFound, "forbidden", nil)
		}

		err = bindFn(ctx, tx, itemID, tagID)
		if err != nil {
			return NewServiceError(ErrItemTagBind, "database error", err)
		}

		err = s.itemRepo.UpdateTx(ctx, tx, item)
		if err != nil {
			return NewServiceError(ErrItemNotUpdated, "database error", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	invalidateSingleCache(ctx, s.cache, itemKey(itemID))
	invalidateListCache(ctx, s.cache, itemListVersionKey(userID))

	return nil
}

func (s *ItemService) ListDeleted(ctx context.Context, f domain.ListItemFilter) ([]domain.Item, int, error) {
	items, count, err := s.itemRepo.ListDeleted(ctx, f)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list deleted items error", err)
	}

	if len(items) > 0 {
		ids := make([]string, len(items))
		for i, item := range items {
			ids[i] = item.ID
		}
		tagsByItem, err := s.tagRepo.FindByItemIDs(ctx, ids)
		if err != nil {
			return nil, 0, NewServiceError(ErrInternal, "get item tags error", err)
		}
		for i := range items {
			items[i].Tags = tagsByItem[items[i].ID]
		}
	}

	return items, count, nil
}

func (s *ItemService) PermanentlyDeleteAllDeleted(ctx context.Context, userID string) error {
	if err := s.itemRepo.PermanentlyDeleteAllDeleted(ctx, userID); err != nil {
		return NewServiceError(ErrInternal, "permanently delete error", err)
	}
	invalidateListCache(ctx, s.cache, itemListVersionKey(userID))
	return nil
}

func (s *ItemService) PermanentlyDeleteByID(ctx context.Context, itemID, userID string) error {
	return s.authorizeAndMutateTx(
		ctx, itemID, userID,
		s.itemRepo.GetDeletedByIDForUpdate,
		s.itemRepo.PermanentlyDeleteByIDTx,
	)
}

func itemKey(itemID string) string {
	return fmt.Sprintf("item:%s", itemID)
}

func itemListVersionKey(userID string) string {
	return fmt.Sprintf("items:version:user:%s", userID)
}

func itemListKey(version int64, f domain.ListItemFilter) string {
	return fmt.Sprintf(
		"items:list:v%d:user:%s:page:%d:size:%d:dir:%s:col:%s:q:%s:tags:%s",
		version, f.UserID, f.Page, f.PageSize, f.Direction, f.Column, f.Query,
		strings.Join(f.TagIDs, ","),
	)
}
