package services

import (
	"context"
	"qvarkk/kvault/internal/domain"
)

type TagRepo interface {
	List(context.Context, domain.ListTagParams) ([]domain.Tag, int, error)
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

func (s *TagService) List(ctx context.Context, params domain.ListTagParams) ([]domain.Tag, int, error) {
	tags, count, err := s.tagRepo.List(ctx, params)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list tags", err)
	}
	return tags, count, nil
}
