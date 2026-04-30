package handlers

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type ItemResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toItemResponse(item *domain.Item) ItemResponse {
	return ItemResponse{
		ID:        item.ID,
		UserID:    item.UserID,
		Type:      string(item.Type),
		Title:     item.Title,
		Content:   item.Content.String,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
}
