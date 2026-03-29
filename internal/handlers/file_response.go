package handlers

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type FileResponse struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func toFileResponse(fileMeta *domain.FileMeta) FileResponse {
	return FileResponse{
		ID:        fileMeta.ID,
		Path:      fileMeta.Path,
		Size:      fileMeta.Size,
		MimeType:  fileMeta.MimeType,
		Status:    string(fileMeta.Status),
		CreatedAt: fileMeta.CreatedAt.Format(time.RFC3339),
	}
}
