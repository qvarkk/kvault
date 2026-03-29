package handlers

import (
	"qvarkk/kvault/internal/domain"
)

type TaskResponse struct {
	Item ItemResponse `json:"item"`
	File FileResponse `json:"file"`
}

func toTaskResponse(item *domain.Item, fileMeta *domain.FileMeta) TaskResponse {
	return TaskResponse{
		Item: toItemResponse(item),
		File: toFileResponse(fileMeta),
	}
}
