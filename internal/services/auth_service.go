package services

import (
	"context"
	"errors"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repo"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUserRepo interface {
	CreateNew(context.Context, *domain.User) error
	GetByID(context.Context, string) (*domain.User, error)
	GetByEmail(context.Context, string) (*domain.User, error)
	UpdateApiKey(ctx context.Context, userID string, apiKey string) (*domain.User, error)
	IsApiKeyUnique(context.Context, string) (bool, error)
}

type AuthService struct {
	userRepo AuthUserRepo
}

func NewAuthService(userRepo AuthUserRepo) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (a *AuthService) GenerateApiKey(ctx context.Context) (string, error) {
	var apiKey string
	for {
		apiKey = GenerateUuidV4()

		isKeyUnique, err := a.userRepo.IsApiKeyUnique(ctx, apiKey)
		if err != nil {
			errMsg := fmt.Sprintf("couldn't verify if api key %s is unique", apiKey)
			return "", NewServiceError(ErrInternal, errMsg, err)
		}

		if isKeyUnique {
			break
		}
	}

	return apiKey, nil
}

func (a *AuthService) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	apiKey, err := a.GenerateApiKey(ctx)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to generate api key", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to hash password", err)
	}

	user := &domain.User{
		Email:    email,
		Password: string(passwordHash),
		APIKey:   apiKey,
	}

	err = a.userRepo.CreateNew(ctx, user)
	if err != nil {
		return nil, NewServiceError(ErrUserNotCreated, "failed to create user", err)
	}

	return user, nil
}

func (a *AuthService) VerifyCredentials(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		errMsg := fmt.Sprintf("failed to find user %s", email)
		if errors.Is(err, repo.ErrNotFound) {
			return nil, NewServiceError(ErrUserNotFound, errMsg, err)
		}
		return nil, NewServiceError(ErrInternal, errMsg, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to compare password hashes", err)
	}

	return user, nil
}

func (a *AuthService) RotateApiKey(
	ctx context.Context,
	userID string,
) (*domain.User, error) {
	apiKey, err := a.GenerateApiKey(ctx)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to generate api key", err)
	}

	user, err := a.userRepo.UpdateApiKey(ctx, userID, apiKey)
	if err != nil {
		errMsg := fmt.Sprintf("failed to update api key. ID: %s; key: %s", userID, apiKey)
		return nil, NewServiceError(ErrInternal, errMsg, err)
	}

	return user, nil
}

func GenerateUuidV4() string {
	return uuid.New().String()
}
