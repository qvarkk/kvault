package services

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repo"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateNew(context.Context, *domain.User) error
	GetById(context.Context, string) (*domain.User, error)
	GetByEmail(context.Context, string) (*domain.User, error)
	GetByApiKey(context.Context, string) (*domain.User, error)
	UpdateApiKey(ctx context.Context, userID string, apiKey string) (*domain.User, error)
	IsApiKeyUnique(context.Context, string) (bool, error)
}

type AuthService struct {
	userRepo *repo.UserRepo
}

func NewAuthService(userRepo *repo.UserRepo) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (a *AuthService) GenerateApiKey(ctx context.Context) (string, error) {
	var apiKey string
	for {
		apiKey = GenerateUuidV4()

		isKeyUnique, err := a.userRepo.IsAPIKeyUnique(ctx, apiKey)
		if err != nil {
			return "", fmt.Errorf(
				"couldn't verify if api key %s is unique: %w",
				apiKey, wrapInternalError(err),
			)
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
		return nil, fmt.Errorf("failed to generate api key: %w", wrapInternalError(err))
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", wrapInternalError(err))
	}

	user := &domain.User{
		Email:    email,
		Password: string(passwordHash),
		APIKey:   apiKey,
	}

	err = a.userRepo.CreateNew(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", wrapDatabaseError(err))
	}

	return user, nil
}

func (a *AuthService) VerifyCredentials(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, fmt.Errorf("failed to find user %s: %w", email, wrapInternalError(err))
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to compare password hashes: %w", wrapInternalError(err))
	}

	return user, nil
}

func (a *AuthService) RotateApiKey(
	ctx context.Context,
	userID string,
) (*domain.User, error) {
	apiKey, err := a.GenerateApiKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", wrapInternalError(err))
	}

	user, err := a.userRepo.UpdateApiKey(ctx, userID, apiKey)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to update api key. ID: %s. Key: %s. Err: %w",
			userID, apiKey, wrapInternalError(err),
		)
	}

	return user, nil
}

func (a *AuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return a.userRepo.GetByID(ctx, userID)
}

func GenerateUuidV4() string {
	return uuid.New().String()
}
