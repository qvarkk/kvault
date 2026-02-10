package services

import (
	"context"
	"qvarkk/kvault/internal/domain"
)

type ItemRepo interface {
	CreateNew(context.Context, *domain.Item) error
}

type ItemService struct {
	itemRepo ItemRepo
}

type CreateItemInput struct {
	UserID     string
	Type       string
	Title      string
	Content    string
	FileMetaID string
}

func NewItemService(itemRepo ItemRepo) *ItemService {
	return &ItemService{itemRepo: itemRepo}
}

func (i *ItemService) CreateNew(ctx context.Context, input CreateItemInput) (*domain.Item, error) {
	item := &domain.Item{
		UserID:     input.UserID,
		Type:       domain.ItemType(input.Type),
		Title:      input.Title,
		Content:    NewNullString(input.Content),
		FileMetaID: NewNullString(input.FileMetaID),
	}

	err := i.itemRepo.CreateNew(ctx, item)
	if err != nil {
		return nil, NewServiceError(ErrItemNotCreated, "database error", err)
	}

	return item, nil
}
