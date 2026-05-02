package services

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"

	"github.com/jmoiron/sqlx"
)

type StopwordRepo interface {
	CreateNew(context.Context, *domain.Stopword) error
	List(context.Context, domain.ListStopwordParams) ([]domain.Stopword, error)
	GetForUpdate(ctx context.Context, tx *sqlx.Tx, word, userID string) (*domain.Stopword, error)
	EnableTx(context.Context, *sqlx.Tx, *domain.Stopword) error
	DisableTx(context.Context, *sqlx.Tx, *domain.Stopword) error
}

type StopwordService struct {
	stopwordRepo StopwordRepo
	transactor   Transactor
}

type CreateStopwordInput struct {
	UserID string
	Word   string
}

func NewStopwordService(stopwordRepo StopwordRepo, transactor Transactor) *StopwordService {
	return &StopwordService{
		stopwordRepo: stopwordRepo,
		transactor:   transactor,
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

	return stopword, nil
}

func (s *StopwordService) List(ctx context.Context, params domain.ListStopwordParams) ([]domain.Stopword, error) {
	stopwords, err := s.stopwordRepo.List(ctx, params)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "list stopwords internal error", err)
	}
	return stopwords, nil
}

func (s *StopwordService) Enable(ctx context.Context, word string, userID string) error {
	return s.authorizeAndMutateTx(ctx, word, userID, s.stopwordRepo.EnableTx)
}

func (s *StopwordService) Disable(ctx context.Context, word string, userID string) error {
	return s.authorizeAndMutateTx(ctx, word, userID, s.stopwordRepo.DisableTx)
}

func (s *StopwordService) authorizeAndMutateTx(
	ctx context.Context, word, userID string,
	mutateFn func(context.Context, *sqlx.Tx, *domain.Stopword) error,
) error {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		stopword, err := s.stopwordRepo.GetForUpdate(ctx, tx, word, userID)
		if err != nil {
			return NewServiceError(ErrStopwordNotFound, "not found", err)
		}

		err = mutateFn(ctx, tx, stopword)
		if err != nil {
			return NewServiceError(ErrInternal, "mutate stopword internal error", err)
		}

		return nil
	})

	return err
}
