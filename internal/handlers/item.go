package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repo"

	"github.com/gin-gonic/gin"
)

type ItemHandler struct {
	itemRepo *repo.ItemRepo
}

func NewItemHandler(itemRepo *repo.ItemRepo) *ItemHandler {
	return &ItemHandler{itemRepo: itemRepo}
}

type createItemRequest struct {
	Type       string `json:"type" binding:"required,oneof=text file url"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	FileMetaID string `json:"file_meta_id" binding:"uuid4"`
}

func (h *ItemHandler) CreateItem(c *gin.Context) {
	var req createItemRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(c, err)
		return
	}

	item := domain.Item{
		Type:       domain.ItemType(req.Type),
		Title:      req.Title,
		Content:    &req.Content,
		FileMetaID: &req.FileMetaID,
	}

	err := h.itemRepo.CreateItem(context.Background(), &item)
	if err != nil {
		abortOnDbError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}
