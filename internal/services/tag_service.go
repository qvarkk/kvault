package services

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"

	"github.com/jmoiron/sqlx"
)

type TagRepo interface {
	CreateNew(context.Context, *domain.Tag) error
	List(context.Context, domain.ListTagFilter) ([]domain.Tag, int, error)
	GetByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.Tag, error)
	UpdateTx(context.Context, *sqlx.Tx, *domain.Tag) error
}

type TagService struct {
	tagRepo    TagRepo
	transactor Transactor
}

func NewTagService(tagRepo TagRepo, transactor Transactor) *TagService {
	return &TagService{
		tagRepo:    tagRepo,
		transactor: transactor,
	}
}

type CreateTagInput struct {
	UserID string
	Name   string
}

type UpdateTagInput struct {
	UserID string
	TagID  string
	Name   string
}

func (s *TagService) CreateNew(
	ctx context.Context,
	input CreateTagInput,
) (*domain.Tag, error) {
	tag := &domain.Tag{
		UserID: input.UserID,
		Name:   input.Name,
	}

	err := s.tagRepo.CreateNew(ctx, tag)
	if err != nil {
		if errors.Is(err, repositories.ErrAlreadyExists) {
			return nil, NewServiceError(ErrTagAlreadyExists, "tag already exists", err)
		}
		return nil, NewServiceError(ErrTagNotCreated, "database error", err)
	}

	return tag, nil
}

func (s *TagService) List(
	ctx context.Context,
	filter domain.ListTagFilter,
) ([]domain.Tag, int, error) {
	tags, count, err := s.tagRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list tags", err)
	}
	return tags, count, nil
}

func (s *TagService) Update(
	ctx context.Context,
	input UpdateTagInput,
) (*domain.Tag, error) {
	var updated *domain.Tag

	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		tag, err := s.tagRepo.GetByIDForUpdate(ctx, tx, input.TagID)
		if err != nil {
			return NewServiceError(ErrTagNotFound, "not found", err)
		}

		if tag.UserID != input.UserID {
			return NewServiceError(ErrTagNotFound, "forbidden", err)
		}

		tag.Name = input.Name

		if err := s.tagRepo.UpdateTx(ctx, tx, tag); err != nil {
			return NewServiceError(ErrInternal, "update tag internal error", err)
		}

		updated = tag
		return nil
	})

	return updated, err
}
