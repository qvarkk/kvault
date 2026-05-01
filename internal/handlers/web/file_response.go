package web

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type FileResponse struct {
	ID           string `json:"id"`
	S3Key        string `json:"s3_key"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

func toFileResponse(file *domain.File) FileResponse {
	return FileResponse{
		ID:           file.ID,
		S3Key:        file.S3Key,
		OriginalName: file.OriginalName,
		Size:         file.Size,
		MimeType:     file.MimeType,
		Status:       string(file.Status),
		CreatedAt:    file.CreatedAt.Format(time.RFC3339),
	}
}
