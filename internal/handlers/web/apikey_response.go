package web

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type ApiKeyResponse struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Current    bool   `json:"current"`
	LastUsedAt string `json:"last_used_at"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
}

func toApiKeyResponse(key domain.ApiKey, currentKeyID string) ApiKeyResponse {
	return ApiKeyResponse{
		ID:         key.ID,
		Label:      key.Label,
		Current:    key.ID == currentKeyID,
		LastUsedAt: key.LastUsedAt.Format(time.RFC3339),
		CreatedAt:  key.CreatedAt.Format(time.RFC3339),
		ExpiresAt:  key.ExpiresAt.Format(time.RFC3339),
	}
}

func toApiKeyResponses(keys []domain.ApiKey, currentKeyID string) []ApiKeyResponse {
	resp := make([]ApiKeyResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, toApiKeyResponse(k, currentKeyID))
	}
	return resp
}
