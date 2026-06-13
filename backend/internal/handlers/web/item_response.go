package web

import (
	"time"

	"qvarkk/kvault/internal/domain"
)

type ItemResponse struct {
	ID               string   `json:"id"`
	UserID           string   `json:"user_id"`
	Title            string   `json:"title"`
	Content          string   `json:"content"`
	SourceURL        string   `json:"source_url,omitempty"`
	UrlStatus        string   `json:"url_status,omitempty"`
	ExtractedContent string   `json:"extracted_content,omitempty"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
	Tags             []TagRef `json:"tags"`
}

func toItemResponse(item *domain.Item) ItemResponse {
	tags := make([]TagRef, len(item.Tags))
	for i, tag := range item.Tags {
		tags[i] = toTagRef(&tag)
	}

	resp := ItemResponse{
		ID:        item.ID,
		UserID:    item.UserID,
		Title:     item.Title,
		Content:   item.Content.String,
		SourceURL: item.SourceURL.String,
		UrlStatus: item.UrlStatus.String,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
		Tags:      tags,
	}

	return resp
}
