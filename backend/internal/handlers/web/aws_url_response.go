package web

import (
	"time"

	"qvarkk/kvault/internal/domain"
)

type AwsUrlResponse struct {
	Url       string    `json:"url"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	Size      int       `json:"size"`
	ExpiresAt time.Time `json:"expires_at"`
}

func toAwsUrlResponse(p *domain.PresignedURL) AwsUrlResponse {
	return AwsUrlResponse{
		Url:       p.URL,
		Filename:  p.Filename,
		MimeType:  p.MimeType,
		Size:      int(p.Size),
		ExpiresAt: p.ExpiresAt,
	}
}
