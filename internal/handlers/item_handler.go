package handlers

import (
	"context"
	"errors"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type ItemService interface {
	CreateNew(context.Context, services.CreateItemInput) (*domain.Item, error)
}

type ItemHandler struct {
	itemService ItemService
}

func NewItemHandler(itemService ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

type createItemRequest struct {
	Type       string `json:"type" binding:"required,oneof=text file url"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	FileMetaID string `json:"file_meta_id" binding:"omitempty,uuid4"`
}

func (h *ItemHandler) Create(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	var req createItemRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	itemInput := services.CreateItemInput{
		UserID:     userID,
		Type:       req.Type,
		Title:      req.Title,
		Content:    req.Content,
		FileMetaID: req.FileMetaID,
	}

	item, err := h.itemService.CreateNew(ctx.Request.Context(), itemInput)
	if errors.Is(err, services.ErrItemNotCreated) || item == nil {
		abortOnDbError(ctx, err)
		return
	} else if errors.Is(err, services.ErrInternal) {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toItemResponse(item))
}
