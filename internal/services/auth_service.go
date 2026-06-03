package services

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"
	"time"

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
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	DeleteByID(ctx context.Context, userID string) error
}

type AuthApiKeyRepo interface {
	Create(ctx context.Context, key *domain.ApiKey) error
	IsHashUnique(ctx context.Context, keyHash string) (bool, error)
	ListByUser(ctx context.Context, userID string) ([]domain.ApiKey, error)
	UpdateLabel(ctx context.Context, id, userID, label string) error
	DeleteByID(ctx context.Context, id, userID string) error
	DeleteByUserExcept(ctx context.Context, userID, keepID string) error
}

type AuthService struct {
	userRepo AuthUserRepo
	keyRepo  AuthApiKeyRepo
	keyTtl   time.Duration
}

func NewAuthService(userRepo AuthUserRepo, keyRepo AuthApiKeyRepo, keyTtl time.Duration) *AuthService {
	return &AuthService{userRepo: userRepo, keyRepo: keyRepo, keyTtl: keyTtl}
}

// GenerateApiKey produces a fresh CSPRNG API key and returns the plaintext (to
// be shown to the caller once) along with the hash that was verified unique.
func (a *AuthService) GenerateApiKey(ctx context.Context) (plaintext, hash string, err error) {
	for {
		plaintext, hash, err = GenerateApiKey()
		if err != nil {
			return "", "", NewServiceError(ErrInternal, "failed to generate api key", err)
		}

		isKeyUnique, err := a.keyRepo.IsHashUnique(ctx, hash)
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
	label string,
) (*domain.User, string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", NewServiceError(ErrInternal, "failed to hash password", err)
	}

	user := &domain.User{
		Username: username,
		Password: string(passwordHash),
	}

	err = a.userRepo.CreateNew(ctx, user)
	if err != nil {
		if errors.Is(err, repositories.ErrAlreadyExists) {
			return nil, "", NewServiceError(ErrUserAlreadyExists, "user already exists", err)
		}
		return nil, "", NewServiceError(ErrUserNotCreated, "database error", err)
	}

	plaintextKey, err := a.issueKey(ctx, user.ID, label)
	if err != nil {
		return nil, "", err
	}

	return user, plaintextKey, nil
}

// VerifyCredentials authenticates by password and, on success, issues a fresh
// per-device API key without touching the user's other keys.
func (a *AuthService) VerifyCredentials(
	ctx context.Context,
	username string,
	password string,
	label string,
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

	plaintextKey, err := a.issueKey(ctx, user.ID, label)
	if err != nil {
		return nil, "", err
	}

	return user, plaintextKey, nil
}

// issueKey generates a fresh key and persists it as a new api_keys row.
func (a *AuthService) issueKey(ctx context.Context, userID, label string) (string, error) {
	plaintextKey, keyHash, err := a.GenerateApiKey(ctx)
	if err != nil {
		return "", err
	}

	key := &domain.ApiKey{
		UserID:    userID,
		KeyHash:   keyHash,
		Label:     label,
		ExpiresAt: time.Now().Add(a.keyTtl),
	}
	if err := a.keyRepo.Create(ctx, key); err != nil {
		return "", NewServiceError(ErrInternal, "failed to store api key", err)
	}

	return plaintextKey, nil
}

func (a *AuthService) ListKeys(ctx context.Context, userID string) ([]domain.ApiKey, error) {
	keys, err := a.keyRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to list api keys", err)
	}
	return keys, nil
}

func (a *AuthService) RenameKey(ctx context.Context, userID, keyID, label string) error {
	if err := a.keyRepo.UpdateLabel(ctx, keyID, userID, label); err != nil {
		return NewServiceError(ErrInternal, "failed to rename api key", err)
	}
	return nil
}

func (a *AuthService) DeleteKey(ctx context.Context, userID, keyID string) error {
	if err := a.keyRepo.DeleteByID(ctx, keyID, userID); err != nil {
		return NewServiceError(ErrInternal, "failed to delete api key", err)
	}
	return nil
}

// Logout invalidates the caller's current key server-side.
func (a *AuthService) Logout(ctx context.Context, userID, keyID string) error {
	if err := a.keyRepo.DeleteByID(ctx, keyID, userID); err != nil {
		return NewServiceError(ErrInternal, "failed to log out", err)
	}
	return nil
}

// LogoutOthers invalidates every key for the user except the current one.
func (a *AuthService) LogoutOthers(ctx context.Context, userID, keyID string) error {
	if err := a.keyRepo.DeleteByUserExcept(ctx, userID, keyID); err != nil {
		return NewServiceError(ErrInternal, "failed to log out other devices", err)
	}
	return nil
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
