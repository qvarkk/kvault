package services

import (
	"context"
	"errors"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repo"
)

type UserRepo interface {
	GetByID(context.Context, string) (*domain.User, error)
	GetByEmail(context.Context, string) (*domain.User, error)
	GetByApiKey(context.Context, string) (*domain.User, error)
}

type UserService struct {
	userRepo UserRepo
}

func NewUserService(userRepo UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

func (u *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return u.getByField(ctx, repo.UserFieldID, userID)
}

func (u *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return u.getByField(ctx, repo.UserFieldEmail, email)
}

func (u *UserService) GetByApiKey(ctx context.Context, apiKey string) (*domain.User, error) {
	return u.getByField(ctx, repo.UserFieldApiKey, apiKey)
}

func (u *UserService) getByField(ctx context.Context, field string, value string) (*domain.User, error) {
	var user *domain.User
	var err error
	switch field {
	case repo.UserFieldID:
		user, err = u.userRepo.GetByID(ctx, value)
	case repo.UserFieldEmail:
		user, err = u.userRepo.GetByEmail(ctx, value)
	case repo.UserFieldApiKey:
		user, err = u.userRepo.GetByApiKey(ctx, value)
	}

	if err != nil {
		errMsg := fmt.Sprintf("failed to find user with %s = %s", field, value)
		if errors.Is(err, repo.ErrNotFound) {
			return nil, NewServiceError(ErrUserNotFound, errMsg, err)
		}
		return nil, NewServiceError(ErrInternal, errMsg, err)
	}

	return user, nil
}
