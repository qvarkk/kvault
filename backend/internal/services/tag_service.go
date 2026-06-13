package services

import (
	"context"
	"errors"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type TagRepo interface {
	CreateNew(context.Context, *domain.Tag) error
	List(context.Context, domain.ListTagFilter) ([]domain.Tag, int, error)
	GetByID(context.Context, string) (*domain.Tag, error)
	GetByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.Tag, error)
	UpdateTx(context.Context, *sqlx.Tx, *domain.Tag) error
	DeleteByID(context.Context, string) error
	FindByItemID(context.Context, string) ([]domain.Tag, error)
	FindByItemIDs(context.Context, []string) (repositories.ItemTagsByID, error)
}

type TagService struct {
	tagRepo     TagRepo
	itemRepo    ItemRepo
	transactor  Transactor
	cache       CacheStore
	tagCacheTtl time.Duration
}

func NewTagService(
	tagRepo TagRepo,
	itemRepo ItemRepo,
	transactor Transactor,
	cache CacheStore,
	cacheTtl time.Duration,
) *TagService {
	return &TagService{
		tagRepo:     tagRepo,
		itemRepo:    itemRepo,
		transactor:  transactor,
		cache:       cache,
		tagCacheTtl: cacheTtl,
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

	invalidateListCache(ctx, s.cache, tagListVersionKey(input.UserID))

	return tag, nil
}

func (s *TagService) List(
	ctx context.Context,
	f domain.ListTagFilter,
) ([]domain.Tag, int, error) {
	version := listVersion(ctx, s.cache, tagListVersionKey(f.UserID))
	cacheKey := tagListKey(version, f)

	if cached := getFromCache[cachedList[domain.Tag]](ctx, s.cache, cacheKey); cached != nil {
		return cached.Entities, cached.Count, nil
	}

	tags, count, err := s.tagRepo.List(ctx, f)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list tags", err)
	}

	setToCache(ctx, s.cache, cacheKey, cachedList[domain.Tag]{Entities: tags, Count: count}, s.tagCacheTtl)

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
	if err != nil {
		return nil, err
	}

	s.invalidateTagAndItemCaches(ctx, input.TagID, input.UserID)

	return updated, err
}

func (s *TagService) DeleteByID(
	ctx context.Context,
	tagID, userID string,
	block bool,
) error {
	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		return NewServiceError(ErrTagNotFound, "not found", err)
	}

	if tag.UserID != userID {
		return NewServiceError(ErrTagNotFound, "forbidden", nil)
	}

	err = s.tagRepo.DeleteByID(ctx, tagID)
	if err != nil {
		return NewServiceError(ErrInternal, "delete tag internal error", err)
	}

	s.invalidateTagAndItemCaches(ctx, tagID, userID)

	return nil
}

func (s *TagService) invalidateTagAndItemCaches(ctx context.Context, tagID, userID string) {
	itemIDs, err := s.itemRepo.FindIDsByTagID(ctx, tagID)
	if err != nil {
		zap.L().Warn(
			"failed to fetch items for tag cache invalidation",
			zap.String("user_id", userID),
			zap.String("tag_id", tagID),
			zap.Error(err),
		)
	}
	for _, itemID := range itemIDs {
		invalidateSingleCache(ctx, s.cache, itemKey(itemID))
	}
	invalidateListCache(ctx, s.cache, itemListVersionKey(userID))
	invalidateListCache(ctx, s.cache, tagListVersionKey(userID))
}

func tagListVersionKey(userID string) string {
	return fmt.Sprintf("tags:version:user:%s", userID)
}

func tagListKey(version int64, f domain.ListTagFilter) string {
	return fmt.Sprintf(
		"tags:list:v%d:user:%s:page:%d:size:%d:dir:%s:col:%s:q:%s",
		version, f.UserID, f.Page, f.PageSize, f.Direction, f.Column, f.Query,
	)
}
