package services

import (
	"context"
	"qvarkk/kvault/internal/domain"
)

type StopwordRepo interface {
	List(context.Context, domain.ListStopwordParams) ([]domain.Stopword, error)
}

type StopwordService struct {
	stopwordRepo StopwordRepo
	transactor   Transactor
}

func NewStopwordService(stopwordRepo StopwordRepo, transactor Transactor) *StopwordService {
	return &StopwordService{
		stopwordRepo: stopwordRepo,
		transactor:   transactor,
	}
}

func (s *StopwordService) List(ctx context.Context, params domain.ListStopwordParams) ([]domain.Stopword, error) {
	stopwords, err := s.stopwordRepo.List(ctx, params)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "list stopwords internal error", err)
	}
	return stopwords, nil
}
