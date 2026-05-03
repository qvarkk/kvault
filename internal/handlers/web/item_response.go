package web

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type ItemResponse struct {
	ID        string   `json:"id"`
	UserID    string   `json:"user_id"`
	Type      string   `json:"type"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	Tags      []TagRef `json:"tags"`
}

func toItemResponse(item *domain.Item) ItemResponse {
	tags := make([]TagRef, len(item.Tags))
	for i, tag := range item.Tags {
		tags[i] = toTagRef(&tag)
	}

	return ItemResponse{
		ID:        item.ID,
		UserID:    item.UserID,
		Type:      string(item.Type),
		Title:     item.Title,
		Content:   item.Content.String,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
		Tags:      tags,
	}
}
