package web

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type TagResponse struct {
	Name      string `json:"word"`
	UserID    string `json:"user_id"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

func toTagResponse(tag *domain.Tag) TagResponse {
	return TagResponse{
		Name:      tag.Name,
		UserID:    tag.UserID,
		UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
		CreatedAt: tag.CreatedAt.Format(time.RFC3339),
	}
}
