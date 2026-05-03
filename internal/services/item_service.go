package services

import (
	"context"
	"qvarkk/kvault/internal/domain"

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
}

type ItemService struct {
	itemRepo   ItemRepo
	tagRepo    TagRepo
	transactor Transactor
}

type CreateItemInput struct {
	UserID  string
	Type    string
	Title   string
	Content string
}

type UpdateItemInput struct {
	ItemID  string
	UserID  string
	Title   *string
	Content *string
}

func NewItemService(itemRepo ItemRepo, tagRepo TagRepo, transactor Transactor) *ItemService {
	return &ItemService{
		itemRepo:   itemRepo,
		tagRepo:    tagRepo,
		transactor: transactor,
	}
}

func (s *ItemService) CreateNew(ctx context.Context, input CreateItemInput) (*domain.Item, error) {
	item := &domain.Item{
		UserID:  input.UserID,
		Type:    domain.ItemType(input.Type),
		Title:   input.Title,
		Content: NewNullString(input.Content),
	}

	err := s.itemRepo.CreateNew(ctx, item)
	if err != nil {
		return nil, NewServiceError(ErrItemNotCreated, "database error", err)
	}

	return item, nil
}

func (s *ItemService) List(ctx context.Context, params domain.ListItemFilter) ([]domain.Item, int, error) {
	items, count, err := s.itemRepo.List(ctx, params)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list items internal error", err)
	}
	return items, count, nil
}

func (s *ItemService) GetByID(ctx context.Context, itemID, userID string) (*domain.Item, error) {
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, NewServiceError(ErrItemNotFound, "not found", err)
	}

	if item.UserID != userID {
		return nil, NewServiceError(ErrItemNotFound, "forbidden", nil)
	}

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

	return updated, err
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

	return err
}

func (s *ItemService) BindTagByItemID(
	ctx context.Context,
	itemID, tagID, userID string,
) error {
	return s.authorizeAndBindTagTx(ctx, itemID, tagID, userID, s.itemRepo.BindTagByItemIDTx)
}

func (s *ItemService) UnbindTagByItemID(
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
	return s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
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
}
