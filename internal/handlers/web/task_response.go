package web

import (
	"qvarkk/kvault/internal/domain"
)

type TaskResponse struct {
	Item ItemResponse `json:"item"`
	File FileResponse `json:"file"`
}

func toTaskResponse(item *domain.Item, fileMeta *domain.File) TaskResponse {
	return TaskResponse{
		Item: toItemResponse(item),
		File: toFileResponse(fileMeta),
	}
}
