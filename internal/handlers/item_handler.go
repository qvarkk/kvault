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
	return &ItemHandler{
		itemService: itemService,
	}
}

type createItemRequest struct {
	Type    string `json:"type" binding:"required,oneof=text url"`
	Title   string `json:"title" binding:"required" example:"Example title"`
	Content string `json:"content" example:"Some content blah blah."`
}

// @Summary      Create an item
// @Description  Creates an item with data passed through body
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body createItemRequest true "Item data"
// @Success      201   {object}  ItemResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items [post]
func (i *ItemHandler) Create(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	var req createItemRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	itemInput := services.CreateItemInput{
		UserID:  userID,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
	}

	item, err := i.itemService.CreateNew(ctx.Request.Context(), itemInput)
	if errors.Is(err, services.ErrItemNotCreated) || item == nil {
		abortOnDbError(ctx, err)
		return
	} else if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toItemResponse(item))
}
