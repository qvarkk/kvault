package web

import (
	"qvarkk/kvault/internal/domain"
	"time"
)

type StopwordResponse struct {
	Word      string `json:"word"`
	Source    string `json:"source"`
	IsEnabled bool   `json:"is_enabled"`
	UpdatedAt string `json:"updated_at"`
}

func toStopwordResponse(stopword *domain.Stopword) StopwordResponse {
	return StopwordResponse{
		Word:      stopword.Word,
		Source:    string(stopword.Source),
		IsEnabled: stopword.IsEnabled,
		UpdatedAt: stopword.UpdatedAt.Format(time.RFC3339),
	}
}
