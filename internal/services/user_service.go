package services

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repositories"
	"time"
)

// touchThreshold throttles last_used_at/expires_at writes: the sliding window is
// only advanced when the key has not been touched within this interval, so a
// burst of requests does not write on every call.
const touchThreshold = time.Minute

type UserRepo interface {
	GetByID(context.Context, string) (*domain.User, error)
}

type UserApiKeyRepo interface {
	GetByHash(ctx context.Context, keyHash string) (*domain.ApiKey, error)
	Touch(ctx context.Context, id string, lastUsed, expires time.Time) error
	DeleteByID(ctx context.Context, id, userID string) error
}

type UserService struct {
	userRepo UserRepo
	keyRepo  UserApiKeyRepo
	keyTtl   time.Duration
}

func NewUserService(userRepo UserRepo, keyRepo UserApiKeyRepo, keyTtl time.Duration) *UserService {
	return &UserService{userRepo: userRepo, keyRepo: keyRepo, keyTtl: keyTtl}
}

// Authenticate resolves a plaintext API key to its user and key id, enforcing
// expiry lazily and sliding the key's TTL forward on use. The returned key id
// lets handlers act on the caller's current device (logout, rotate).
func (u *UserService) Authenticate(ctx context.Context, apiKey string) (*domain.User, string, error) {
	key, err := u.keyRepo.GetByHash(ctx, HashApiKey(apiKey))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, "", NewServiceError(ErrUnauthenticated, "invalid api key", err)
		}
		return nil, "", NewServiceError(ErrInternal, "failed to look up api key", err)
	}

	now := time.Now()
	if !key.ExpiresAt.After(now) {
		// Expired: invalidate lazily and reject as if the key never existed.
		_ = u.keyRepo.DeleteByID(ctx, key.ID, key.UserID)
		return nil, "", NewServiceError(ErrUnauthenticated, "api key expired", nil)
	}

	if now.Sub(key.LastUsedAt) > touchThreshold {
		_ = u.keyRepo.Touch(ctx, key.ID, now, now.Add(u.keyTtl))
	}

	user, err := u.userRepo.GetByID(ctx, key.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, "", NewServiceError(ErrUnauthenticated, "invalid api key", err)
		}
		return nil, "", NewServiceError(ErrInternal, "failed to look up user", err)
	}

	return user, key.ID, nil
}

func (u *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, NewServiceError(ErrUserNotFound, "user not found", err)
		}
		return nil, NewServiceError(ErrInternal, "failed to look up user", err)
	}
	return user, nil
}
