package domain

import "time"

type PresignedURL struct {
	URL       string
	Filename  string
	MimeType  string
	Size      int64
	ExpiresAt time.Time
}
