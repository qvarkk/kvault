package services

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"
)

type TagRepo interface {
	CreateNew(context.Context, *domain.Tag) error
	List(context.Context, domain.ListTagFilter) ([]domain.Tag, int, error)
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
