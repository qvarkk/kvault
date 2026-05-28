package web

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	APIKey    string `json:"api_key,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toUserResponseWithApiKey(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		APIKey:    user.APIKey,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
