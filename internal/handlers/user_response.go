package handlers

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	APIKey    string `json:"api_key,omitempty"`
	CreatedAt string `json:"created_at"`
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}

func toUserResponseWithApiKey(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		APIKey:    user.APIKey,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
