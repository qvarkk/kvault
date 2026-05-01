package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type ItemService interface {
	CreateNew(context.Context, services.CreateItemInput) (*domain.Item, error)
	List(context.Context, domain.ListItemParams) ([]domain.Item, int, error)
	GetByID(ctx context.Context, itemID, userID string) (*domain.Item, error)
	DeleteByID(ctx context.Context, itemID, userID string) error
	Update(context.Context, services.UpdateItemInput) (*domain.Item, error)
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

type itemIDUri struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type UpdateItemRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
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
func (h *ItemHandler) Create(ctx *gin.Context) error {
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

	item, err := h.itemService.CreateNew(ctx.Request.Context(), itemInput)
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
func (h *ItemHandler) List(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req listItemRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return err
	}

	params := domain.ListItemParams{
		UserID:    userID,
		Query:     req.Query,
		Type:      req.Type,
		Page:      req.Page,
		PageSize:  req.PageSize,
		Direction: req.Direction,
		Column:    req.Column,
	}

	items, total, err := h.itemService.List(ctx, params)
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
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/{id} [get]
func (h *ItemHandler) Get(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri itemIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	item, err := h.itemService.GetByID(ctx.Request.Context(), uri.ID, userID)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toItemResponse(item))
	return nil
}

// @Summary      Soft delete an item
// @Description  Marks an item with given ID as deleted if it's owned by the User
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Item ID"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/{id} [delete]
func (h *ItemHandler) Delete(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri itemIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	err := h.itemService.DeleteByID(ctx.Request.Context(), uri.ID, userID)
	if err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Update an item in your vault
// @Description  Partially updates an item by ID
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string            true  "Item ID"
// @Param        body body      UpdateItemRequest  true  "Fields to update"
// @Success      200  {object}  ItemResponse
// @Failure      401  {object}  httpx.ErrorResponse
// @Failure      404  {object}  httpx.ErrorResponse
// @Failure      422  {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500  {object}  httpx.ErrorResponse
// @Router       /items/{id} [patch]
func (h *ItemHandler) Update(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri itemIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	var req UpdateItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return err
	}

	itemInput := services.UpdateItemInput{
		ItemID:  uri.ID,
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}

	item, err := h.itemService.Update(ctx.Request.Context(), itemInput)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toItemResponse(item))
	return nil
}
