package services

import (
	"context"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"
)

type ItemRepo interface {
	CreateNew(context.Context, *domain.Item) error
	List(context.Context, repositories.ListItemParams) ([]domain.Item, int, error)
	GetByID(context.Context, string) (*domain.Item, error)
	SoftDeleteByID(context.Context, string) error
}

type ItemService struct {
	itemRepo ItemRepo
}

type CreateItemInput struct {
	UserID  string
	Type    string
	Title   string
	Content string
}

type ListItemParams repositories.ListItemParams

func NewItemService(itemRepo ItemRepo) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
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

func (s *ItemService) List(ctx context.Context, params ListItemParams) ([]domain.Item, int, error) {
	items, count, err := s.itemRepo.List(ctx, repositories.ListItemParams(params))
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

func (s *ItemService) DeleteByID(ctx context.Context, itemID, userID string) error {
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return NewServiceError(ErrItemNotFound, "not found", err)
	}

	if item.UserID != userID {
		return NewServiceError(ErrItemNotFound, "forbidden", nil)
	}

	err = s.itemRepo.SoftDeleteByID(ctx, itemID)
	if err != nil {
		return NewServiceError(ErrInternal, "delete item internal error", err)
	}

	return nil
}
