package web

import (
	"encoding/json"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"
	"time"
)

type ItemResponse struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	Type             string          `json:"type"`
	Title            string          `json:"title"`
	Content          string          `json:"content"`
	SourceURL        string          `json:"source_url,omitempty"`
	UrlMetadata      *UrlMetadataDTO `json:"url_metadata,omitempty"`
	ExtractedContent string          `json:"extracted_content,omitempty"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
	Tags             []TagRef        `json:"tags"`
}

type UrlMetadataDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
}

func toItemResponse(item *domain.Item) ItemResponse {
	tags := make([]TagRef, len(item.Tags))
	for i, tag := range item.Tags {
		tags[i] = toTagRef(&tag)
	}

	resp := ItemResponse{
		ID:        item.ID,
		UserID:    item.UserID,
		Type:      string(item.Type),
		Title:     item.Title,
		Content:   item.Content.String,
		SourceURL: item.SourceURL.String,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
		Tags:      tags,
	}

	if item.UrlMetadata.Valid && item.UrlMetadata.String != "" {
		var meta services.UrlMetadata
		if err := json.Unmarshal([]byte(item.UrlMetadata.String), &meta); err == nil {
			resp.UrlMetadata = &UrlMetadataDTO{
				Title:       meta.Title,
				Description: meta.Description,
				Image:       meta.Image,
				SiteName:    meta.SiteName,
			}
		}
	}

	if item.ExtractedContent.Valid {
		resp.ExtractedContent = item.ExtractedContent.String
	}

	return resp
}
