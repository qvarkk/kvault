package web

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type TagRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TagResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UserID    string `json:"user_id"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

func toTagResponse(tag *domain.Tag) TagResponse {
	return TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		UserID:    tag.UserID,
		UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
		CreatedAt: tag.CreatedAt.Format(time.RFC3339),
	}
}

func toTagRef(tag *domain.Tag) TagRef {
	return TagRef{
		ID:   tag.ID,
		Name: tag.Name,
	}
}
