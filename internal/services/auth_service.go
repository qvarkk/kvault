package services

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

// dummyPasswordHash is compared against on the user-not-found login path so the
// response time does not leak whether a username exists.
var dummyPasswordHash []byte

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("kvault-dummy-password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	dummyPasswordHash = h
}

type AuthUserRepo interface {
	CreateNew(context.Context, *domain.User) error
	GetByID(context.Context, string) (*domain.User, error)
	GetByUsername(context.Context, string) (*domain.User, error)
	UpdateApiKey(ctx context.Context, userID string, apiKey string) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	DeleteByID(ctx context.Context, userID string) error
	IsApiKeyUnique(context.Context, string) (bool, error)
}

type AuthService struct {
	userRepo AuthUserRepo
}

func NewAuthService(userRepo AuthUserRepo) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// GenerateApiKey produces a fresh CSPRNG API key and returns the plaintext (to
// be shown to the caller once) along with the hash that was verified unique.
func (a *AuthService) GenerateApiKey(ctx context.Context) (plaintext, hash string, err error) {
	for {
		plaintext, hash, err = GenerateApiKey()
		if err != nil {
			return "", "", NewServiceError(ErrInternal, "failed to generate api key", err)
		}

		isKeyUnique, err := a.userRepo.IsApiKeyUnique(ctx, hash)
		if err != nil {
			return "", "", NewServiceError(ErrInternal, "couldn't verify api key uniqueness", err)
		}

		if isKeyUnique {
			return plaintext, hash, nil
		}
	}
}

func (a *AuthService) RegisterNewUser(
	ctx context.Context,
	username string,
	password string,
) (*domain.User, string, error) {
	plaintextKey, keyHash, err := a.GenerateApiKey(ctx)
	if err != nil {
		return nil, "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", NewServiceError(ErrInternal, "failed to hash password", err)
	}

	user := &domain.User{
		Username:   username,
		Password:   string(passwordHash),
		APIKeyHash: keyHash,
	}

	err = a.userRepo.CreateNew(ctx, user)
	if err != nil {
		if errors.Is(err, repositories.ErrAlreadyExists) {
			return nil, "", NewServiceError(ErrUserAlreadyExists, "user already exists", err)
		}
		return nil, "", NewServiceError(ErrUserNotCreated, "database error", err)
	}

	return user, plaintextKey, nil
}

// VerifyCredentials authenticates by password and, on success, rotates the
// user's API key (rotate-on-login), returning the fresh plaintext key.
func (a *AuthService) VerifyCredentials(
	ctx context.Context,
	username string,
	password string,
) (*domain.User, string, error) {
	user, err := a.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			// Equalize timing with the wrong-password path so the response time
			// does not reveal whether the username exists.
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
			return nil, "", NewServiceError(ErrInvalidCredentials, "user not found", err)
		}
		return nil, "", NewServiceError(ErrInternal, "failed to look up user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", NewServiceError(ErrInvalidCredentials, "wrong password", err)
	}

	return a.issueNewKey(ctx, user.ID)
}

func (a *AuthService) RotateApiKey(
	ctx context.Context,
	userID string,
) (*domain.User, string, error) {
	return a.issueNewKey(ctx, userID)
}

// issueNewKey generates a fresh API key, persists its hash for the given user,
// and returns the updated user together with the plaintext key to show once.
func (a *AuthService) issueNewKey(ctx context.Context, userID string) (*domain.User, string, error) {
	plaintextKey, keyHash, err := a.GenerateApiKey(ctx)
	if err != nil {
		return nil, "", err
	}

	user, err := a.userRepo.UpdateApiKey(ctx, userID, keyHash)
	if err != nil {
		return nil, "", NewServiceError(ErrInternal, "failed to update api key", err)
	}

	return user, plaintextKey, nil
}

func (a *AuthService) ChangePassword(
	ctx context.Context,
	userID, oldPassword, newPassword string,
) error {
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return NewServiceError(ErrUserNotFound, "not found", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return NewServiceError(ErrInvalidCredentials, "wrong old password", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return NewServiceError(ErrInternal, "failed to hash password", err)
	}

	if err := a.userRepo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return NewServiceError(ErrInternal, "failed to update password", err)
	}

	return nil
}

func (a *AuthService) VerifyPassword(ctx context.Context, userID, password string) error {
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return NewServiceError(ErrUserNotFound, "not found", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return NewServiceError(ErrInvalidCredentials, "wrong password", err)
	}
	return nil
}

func (a *AuthService) DeleteAccount(ctx context.Context, userID string) error {
	if err := a.userRepo.DeleteByID(ctx, userID); err != nil {
		return NewServiceError(ErrInternal, "failed to delete user", err)
	}
	return nil
}
