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
	List(context.Context, domain.ListItemFilter) ([]domain.Item, int, error)
	ListDeleted(context.Context, domain.ListItemFilter) ([]domain.Item, int, error)
	GetByID(ctx context.Context, itemID, userID string) (*domain.Item, error)
	DeleteByID(ctx context.Context, itemID, userID string) error
	PermanentlyDeleteAllDeleted(ctx context.Context, userID string) error
	PermanentlyDeleteByID(ctx context.Context, itemID, userID string) error
	Update(context.Context, services.UpdateItemInput) (*domain.Item, error)
	RestoreByID(ctx context.Context, itemID, userID string) error
	AttachTagByItemID(ctx context.Context, itemID, tagID, userID string) error
	DetachTagByItemID(ctx context.Context, itemID, tagID, userID string) error
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
	Title     string `json:"title" binding:"required" example:"Example title"`
	Content   string `json:"content" example:"Some content blah blah."`
	SourceURL string `json:"source_url" binding:"omitempty,url" example:"https://example.com/article"`
}

type listItemQuery struct {
	Query  string   `form:"q"`
	TagIDs []string `form:"tag_ids" binding:"omitempty,dive,uuid" collectionFormat:"multi"`
	PaginationParams
	ItemSortingParams
}

type itemIDUri struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type updateItemRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

type bindTagRequest struct {
	TagID string `json:"tag_id" binding:"required,uuid"`
}

type autotagRequest struct {
	Number int `json:"number" binding:"required,min=1,max=10"`
}

type unbindTagUri struct {
	ItemID string `uri:"id" binding:"required,uuid"`
	TagID  string `uri:"tag_id" binding:"required,uuid"`
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
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		SourceURL: req.SourceURL,
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
// @Param				 params  query listItemQuery false "Query parameters"
// @Success      200   {object}  PaginatedResponse[ItemResponse]
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items [get]
func (h *ItemHandler) List(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var query listItemQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		return err
	}

	params := domain.ListItemFilter{
		UserID: userID,
		TagIDs: query.TagIDs,
		QueryFilter: domain.QueryFilter{
			Query: query.Query,
		},
		PaginationFilter: domain.PaginationFilter{
			Page:     query.Page,
			PageSize: query.PageSize,
		},
		SortFilter: domain.SortFilter{
			Direction: query.Direction,
			Column:    query.Column,
		},
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

// @Summary      Update an item in your vault
// @Description  Partially updates an item by ID
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string            true  "Item ID"
// @Param        body body      updateItemRequest true  "Fields to update"
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

	var req updateItemRequest
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
func (h *ItemHandler) SoftDelete(ctx *gin.Context) error {
	return h.withOwnedItemAction(ctx, h.itemService.DeleteByID)
}

// @Summary      Restore a soft deleted item
// @Description  Unmarks an item with given ID as deleted if it's owned by the User
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
// @Router       /items/{id}/restore [post]
func (h *ItemHandler) Restore(ctx *gin.Context) error {
	return h.withOwnedItemAction(ctx, h.itemService.RestoreByID)
}

func (h *ItemHandler) withOwnedItemAction(
	ctx *gin.Context,
	fn func(context.Context, string, string) error,
) error {
	userID := ctx.MustGet("userID").(string)

	var uri itemIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	err := fn(ctx.Request.Context(), uri.ID, userID)
	if err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Attach a tag to the item
// @Description  Creates a binding between given item and tag
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id     path   string          true  "Item ID"
// @Param        body   body   bindTagRequest  true  "Tag info"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/{id}/tags [post]
func (h *ItemHandler) AttachTag(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri itemIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	var req bindTagRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	err := h.itemService.AttachTagByItemID(ctx.Request.Context(), uri.ID, req.TagID, userID)
	if err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Remove a tag from the item
// @Description  Deletes a binding between given item and tag
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        item_id  path  string  true  "Item ID"
// @Param        tag_id   path  string  true  "Tag ID"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/{item_id}/tags/{tag_id} [delete]
func (h *ItemHandler) DetachTag(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri unbindTagUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	err := h.itemService.DetachTagByItemID(ctx.Request.Context(), uri.ItemID, uri.TagID, userID)
	if err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      List deleted items
// @Description  Returns soft-deleted items owned by the user
// @Tags         Items
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        params query listItemQuery false "Query parameters"
// @Success      200   {object}  PaginatedResponse[ItemResponse]
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/deleted [get]
func (h *ItemHandler) ListDeleted(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var query listItemQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		return err
	}

	params := domain.ListItemFilter{
		UserID: userID,
		PaginationFilter: domain.PaginationFilter{
			Page:     query.Page,
			PageSize: query.PageSize,
		},
		SortFilter: domain.SortFilter{
			Direction: query.Direction,
			Column:    query.Column,
		},
	}

	items, total, err := h.itemService.ListDeleted(ctx, params)
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

// @Summary      Permanently delete a single item from trash
// @Description  Permanently deletes a single soft-deleted item owned by the user
// @Tags         Items
// @Security     ApiKeyAuth
// @Param        id path string true "Item ID"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/deleted/{id} [delete]
func (h *ItemHandler) PermanentlyDelete(ctx *gin.Context) error {
	return h.withOwnedItemAction(ctx, h.itemService.PermanentlyDeleteByID)
}

// @Summary      Clear item trash
// @Description  Permanently deletes all soft-deleted items owned by the user
// @Tags         Items
// @Security     ApiKeyAuth
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /items/deleted [delete]
func (h *ItemHandler) ClearSoftDeleted(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	if err := h.itemService.PermanentlyDeleteAllDeleted(ctx.Request.Context(), userID); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}
