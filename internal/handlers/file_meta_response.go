package handlers

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type FileMetaResponse struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

func toFileMetaResponse(item *domain.FileMeta) FileMetaResponse {
	return FileMetaResponse{
		ID:        item.ID,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
}
