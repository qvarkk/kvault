package services

import (
	"context"
	"qvarkk/kvault/internal/domain"
)

type ItemRepo interface {
	CreateNew(context.Context, *domain.Item) error
	GetByID(context.Context, string) (*domain.Item, error)
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

func NewItemService(itemRepo ItemRepo) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
	}
}

func (i *ItemService) CreateNew(ctx context.Context, input CreateItemInput) (*domain.Item, error) {
	item := &domain.Item{
		UserID:  input.UserID,
		Type:    domain.ItemType(input.Type),
		Title:   input.Title,
		Content: NewNullString(input.Content),
	}

	err := i.itemRepo.CreateNew(ctx, item)
	if err != nil {
		return nil, NewServiceError(ErrItemNotCreated, "database error", err)
	}

	return item, nil
}

func (i *ItemService) GetByID(ctx context.Context, itemID, userID string) (*domain.Item, error) {
	item, err := i.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, NewServiceError(ErrItemNotFound, "not found", err)
	}

	if item.UserID != userID {
		return nil, NewServiceError(ErrItemNotFound, "forbidden", nil)
	}

	return item, nil
}
