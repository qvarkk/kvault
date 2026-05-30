package services

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"
)

type UserRepo interface {
	GetByID(context.Context, string) (*domain.User, error)
	GetByApiKeyHash(context.Context, string) (*domain.User, error)
}

type UserService struct {
	userRepo UserRepo
}

func NewUserService(userRepo UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

func (u *UserService) Authenticate(ctx context.Context, apiKey string) (*domain.User, error) {
	user, err := u.GetByApiKeyHash(ctx, HashApiKey(apiKey))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, NewServiceError(ErrUnauthenticated, "invalid api key", err)
		}
		return nil, err
	}

	return user, nil
}

func (u *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return u.getByField(ctx, repositories.UserFieldID, userID)
}

func (u *UserService) GetByApiKeyHash(ctx context.Context, apiKeyHash string) (*domain.User, error) {
	return u.getByField(ctx, repositories.UserFieldApiKey, apiKeyHash)
}

func (u *UserService) getByField(ctx context.Context, field string, value string) (*domain.User, error) {
	var user *domain.User
	var err error
	switch field {
	case repositories.UserFieldID:
		user, err = u.userRepo.GetByID(ctx, value)
	case repositories.UserFieldApiKey:
		user, err = u.userRepo.GetByApiKeyHash(ctx, value)
	}

	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, NewServiceError(ErrUserNotFound, "user not found", err)
		}
		return nil, NewServiceError(ErrInternal, "failed to look up user", err)
	}

	return user, nil
}
