package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type ItemService interface {
	CreateNew(context.Context, services.CreateItemInput) (*domain.Item, error)
	List(context.Context, services.ListItemParams) ([]domain.Item, int, error)
	GetByID(ctx context.Context, itemID, userID string) (*domain.Item, error)
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

type listItemRequest struct {
	Query string `form:"q"`
	Type  string `form:"type" binding:"omitempty,oneof=text url"`
	PaginationParams
	ItemSortingParams
}

type getItemUri struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// @Summary      Create an item in your vault
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
func (i *ItemHandler) Create(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req createItemRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	itemInput := services.CreateItemInput{
		UserID:  userID,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
	}

	item, err := i.itemService.CreateNew(ctx.Request.Context(), itemInput)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusCreated, toItemResponse(item))
	return nil
}

// @Summary      Get all items
// @Description  Returns a list of items owned by the User
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param				 params query listItemRequest false "Query parameters"
// @Success      200   {object}  ItemResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items [get]
func (i *ItemHandler) List(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req listItemRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return err
	}

	params := services.ListItemParams{
		UserID:    userID,
		Query:     req.Query,
		Type:      req.Type,
		Page:      req.Page,
		PageSize:  req.PageSize,
		Direction: req.Direction,
		Column:    req.Column,
	}

	items, total, err := i.itemService.List(ctx, params)
	if err != nil {
		return err
	}

	itemResponses := make([]ItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = toItemResponse(&item)
	}

	ctx.JSON(http.StatusOK, toPaginatedResponse(itemResponses, total, params.Page, params.PageSize))
	return nil
}

// @Summary      Get an item from your vault
// @Description  Gets an item by ID
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Item ID"
// @Success      200   {object}  ItemResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      403   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/{id} [get]
func (i *ItemHandler) Get(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri getItemUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	item, err := i.itemService.GetByID(ctx.Request.Context(), uri.ID, userID)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toItemResponse(item))
	return nil
}
