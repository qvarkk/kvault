package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"

	"github.com/gin-gonic/gin"
)

type StopwordService interface {
	List(context.Context, domain.ListStopwordParams) ([]domain.Stopword, error)
}

type StopwordHandler struct {
	stopwordService StopwordService
}

func NewStopwordHandler(stopwordService StopwordService) *StopwordHandler {
	return &StopwordHandler{stopwordService: stopwordService}
}

type listStopwordRequest struct {
	Query  string `form:"q"`
	Source string `form:"source" binding:"omitempty,oneof=user default"`
	StopwordSortingParams
}

// @Summary      Get all stopwords
// @Description  Returns a list of stopwords used by the User
// @Tags         Stopwords
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param				 params query listStopwordRequest false "Query parameters"
// @Success      200   {object}  ListResponse[StopwordResponse]
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /stopwords [get]
func (h *StopwordHandler) List(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req listStopwordRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return err
	}

	params := domain.ListStopwordParams{
		UserID:    userID,
		Query:     req.Query,
		Source:    req.Source,
		Direction: req.Direction,
		Column:    req.Column,
	}

	stopwords, err := h.stopwordService.List(ctx, params)
	if err != nil {
		return err
	}

	stopwordResponses := make([]StopwordResponse, len(stopwords))
	for i, stopword := range stopwords {
		stopwordResponses[i] = toStopwordResponse(&stopword)
	}

	ctx.JSON(http.StatusOK, toListResponse(stopwordResponses))
	return nil
}
