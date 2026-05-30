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

// toUserResponse renders a user without the API key (for /me and any read).
func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

// toUserResponseWithApiKey includes the freshly issued plaintext key. Used only
// by register, login (rotate-on-login), and refresh — never by /me — because the
// key is stored hashed and cannot be recovered afterwards.
func toUserResponseWithApiKey(user *domain.User, apiKey string) UserResponse {
	resp := toUserResponse(user)
	resp.APIKey = apiKey
	return resp
}
