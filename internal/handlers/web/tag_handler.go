package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"

	"github.com/gin-gonic/gin"
)

type TagService interface {
	List(context.Context, domain.ListTagParams) ([]domain.Tag, int, error)
}

type TagHandler struct {
	tagService TagService
}

func NewTagHandler(tagService TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

type listTagRequest struct {
	Query string `form:"q"`
	PaginationParams
	TagSortingParams
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

	params := domain.ListTagParams{
		UserID:    userID,
		Query:     req.Query,
		Page:      req.Page,
		PageSize:  req.PageSize,
		Direction: req.Direction,
		Column:    req.Column,
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
