package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type TagService interface {
	CreateNew(context.Context, services.CreateTagInput) (*domain.Tag, error)
	List(context.Context, domain.ListTagFilter) ([]domain.Tag, int, error)
}

type TagHandler struct {
	tagService TagService
}

func NewTagHandler(tagService TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

type createTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type listTagRequest struct {
	Query string `form:"q"`
	PaginationParams
	TagSortingParams
}

// @Summary      Create a tag in your vault
// @Description  Creates a tag with data passed through body
// @Tags         Tags
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body createTagRequest true "Tag data"
// @Success      201   {object}  TagResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /tags [post]
func (h *TagHandler) Create(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req createTagRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	tagInput := services.CreateTagInput{
		UserID: userID,
		Name:   req.Name,
	}

	tag, err := h.tagService.CreateNew(ctx.Request.Context(), tagInput)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusCreated, toTagResponse(tag))
	return nil
}

// @Summary      Get all tags
// @Description  Returns a paginated list of tags used by the User
// @Tags         Tags
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param				 params query listTagRequest false "Query parameters"
// @Success      200   {object}  PaginatedResponse[TagResponse]
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /tags [get]
func (h *TagHandler) List(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req listTagRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return err
	}

	params := domain.ListTagFilter{
		UserID: userID,
		QueryFilter: domain.QueryFilter{
			Query: req.Query,
		},
		PaginationFilter: domain.PaginationFilter{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		SortFilter: domain.SortFilter{
			Direction: req.Direction,
			Column:    req.Column,
		},
	}

	tags, count, err := h.tagService.List(ctx, params)
	if err != nil {
		return err
	}

	tagResponses := make([]TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = toTagResponse(&tag)
	}

	ctx.JSON(http.StatusOK, toPaginatedResponse(tagResponses, count, req.Page, req.PageSize))
	return nil
}
