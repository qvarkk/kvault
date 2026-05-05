package services

import (
	"context"
	"errors"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"
	"time"

	"github.com/jmoiron/sqlx"
)

type StopwordRepo interface {
	CreateNew(context.Context, *domain.Stopword) error
	UpsertTx(context.Context, *sqlx.Tx, *domain.Stopword) error
	GetActiveStopwords(context.Context, domain.ListStopwordFilter) ([]domain.Stopword, error)
	GetForUpdate(ctx context.Context, tx *sqlx.Tx, word, userID string) (*domain.Stopword, error)
	Get(ctx context.Context, word, userID string) (*domain.Stopword, error)
	EnableTx(ctx context.Context, tx *sqlx.Tx, word, userID string) error
	DisableTx(ctx context.Context, tx *sqlx.Tx, word, userID string) error
	Delete(ctx context.Context, word, userID string) error
	DeleteTx(ctx context.Context, tx *sqlx.Tx, word, userID string) error
	IsDefaultTx(context.Context, *sqlx.Tx, string) (bool, error)
}

type StopwordService struct {
	stopwordRepo     StopwordRepo
	transactor       Transactor
	cache            CacheStore
	stopwordCacheTtl time.Duration
}

type CreateStopwordInput struct {
	UserID string
	Word   string
}

func NewStopwordService(
	stopwordRepo StopwordRepo,
	transactor Transactor,
	cache CacheStore,
	cacheTtl time.Duration,
) *StopwordService {
	return &StopwordService{
		stopwordRepo:     stopwordRepo,
		transactor:       transactor,
		cache:            cache,
		stopwordCacheTtl: cacheTtl,
	}
}

func (s *StopwordService) CreateNew(ctx context.Context, input CreateStopwordInput) (*domain.Stopword, error) {
	stopword := &domain.Stopword{
		UserID:    input.UserID,
		Word:      input.Word,
		Source:    domain.StopwordSourceUser,
		IsEnabled: true,
	}

	err := s.stopwordRepo.CreateNew(ctx, stopword)
	if err != nil {
		if errors.Is(err, repositories.ErrAlreadyExists) {
			return nil, NewServiceError(ErrStopwordAlreadyExists, "stopword already exists", err)
		}
		return nil, NewServiceError(ErrStopwordNotCreated, "database error", err)
	}

	invalidateListCache(ctx, s.cache, stopwordListVersionKey(input.UserID))

	return stopword, nil
}

func (s *StopwordService) List(ctx context.Context, f domain.ListStopwordFilter) ([]domain.Stopword, error) {
	version := listVersion(ctx, s.cache, stopwordListVersionKey(f.UserID))
	cacheKey := stopwordListKey(version, f)

	if cached := getFromCache[[]domain.Stopword](ctx, s.cache, cacheKey); cached != nil {
		return *cached, nil
	}

	stopwords, err := s.stopwordRepo.GetActiveStopwords(ctx, f)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "list stopwords internal error", err)
	}

	setToCache(ctx, s.cache, cacheKey, stopwords, s.stopwordCacheTtl)

	return stopwords, nil
}

func (s *StopwordService) Enable(ctx context.Context, word string, userID string) error {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		isDefault, err := s.stopwordRepo.IsDefaultTx(ctx, tx, word)
		if err != nil {
			return NewServiceError(ErrInternal, "check default", err)
		}

		if isDefault {
			err = s.stopwordRepo.DeleteTx(ctx, tx, word, userID)
			if err != nil {
				return NewServiceError(ErrInternal, "delete default", err)
			}
			return nil
		}

		_, err = s.stopwordRepo.GetForUpdate(ctx, tx, word, userID)
		if err != nil {
			return NewServiceError(ErrStopwordNotFound, "not found", err)
		}

		err = s.stopwordRepo.EnableTx(ctx, tx, word, userID)
		if err != nil {
			return NewServiceError(ErrInternal, "mutate stopword internal error", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	invalidateListCache(ctx, s.cache, stopwordListVersionKey(userID))

	return nil
}

func (s *StopwordService) Disable(ctx context.Context, word string, userID string) error {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		isDefault, err := s.stopwordRepo.IsDefaultTx(ctx, tx, word)
		if err != nil {
			return NewServiceError(ErrInternal, "check default", err)
		}

		if isDefault {
			stopword := &domain.Stopword{
				UserID:    userID,
				Word:      word,
				Source:    domain.StopwordSourceDefault,
				IsEnabled: false,
			}

			err = s.stopwordRepo.UpsertTx(ctx, tx, stopword)
			if err != nil {
				return NewServiceError(ErrInternal, "upsert default", err)
			}
			return nil
		}

		_, err = s.stopwordRepo.GetForUpdate(ctx, tx, word, userID)
		if err != nil {
			return NewServiceError(ErrStopwordNotFound, "not found", err)
		}

		err = s.stopwordRepo.DisableTx(ctx, tx, word, userID)
		if err != nil {
			return NewServiceError(ErrInternal, "mutate stopword internal error", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	invalidateListCache(ctx, s.cache, stopwordListVersionKey(userID))

	return nil
}

func (s *StopwordService) Delete(
	ctx context.Context,
	word, userID string,
) error {
	_, err := s.stopwordRepo.Get(ctx, word, userID)
	if err != nil {
		return NewServiceError(ErrStopwordNotFound, "not found", err)
	}

	err = s.stopwordRepo.Delete(ctx, word, userID)
	if err != nil {
		return NewServiceError(ErrInternal, "delete stopword internal error", err)
	}

	invalidateListCache(ctx, s.cache, stopwordListVersionKey(userID))

	return nil
}

func stopwordListVersionKey(userID string) string {
	return fmt.Sprintf("stopwords:version:user:%s", userID)
}

func stopwordListKey(version int64, f domain.ListStopwordFilter) string {
	return fmt.Sprintf(
		"stopwords:list:v%d:user:%s:source:%s:dir:%s:col:%s:q:%s",
		version, f.UserID, f.Source, f.Direction, f.Column, f.Query,
	)
}
