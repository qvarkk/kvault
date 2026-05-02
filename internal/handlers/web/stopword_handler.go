package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type StopwordService interface {
	CreateNew(context.Context, services.CreateStopwordInput) (*domain.Stopword, error)
	List(context.Context, domain.ListStopwordFilter) ([]domain.Stopword, error)
	Enable(ctx context.Context, word, userID string) error
	Disable(ctx context.Context, word, userID string) error
	Delete(ctx context.Context, word, userID string) error
}

type StopwordHandler struct {
	stopwordService StopwordService
}

func NewStopwordHandler(stopwordService StopwordService) *StopwordHandler {
	return &StopwordHandler{stopwordService: stopwordService}
}

type createStopwordRequest struct {
	Word string `json:"word" binding:"required"`
}

type listStopwordRequest struct {
	Query  string `form:"q"`
	Source string `form:"source" binding:"omitempty,oneof=user default"`
	StopwordSortingParams
}

type stopwordUri struct {
	Word string `uri:"word" binding:"required"`
}

// @Summary      Add a stopword to the user profile
// @Description  Creates a stopword record in database
// @Tags         Stopwords
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        body body createStopwordRequest true "Stopword data"
// @Success      201   {object}  StopwordResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /stopwords [post]
func (h *StopwordHandler) Create(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req createStopwordRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	stopwordInput := services.CreateStopwordInput{
		UserID: userID,
		Word:   req.Word,
	}

	stopword, err := h.stopwordService.CreateNew(ctx.Request.Context(), stopwordInput)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusCreated, toStopwordResponse(stopword))
	return nil
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

	params := domain.ListStopwordFilter{
		UserID: userID,
		Source: req.Source,
		QueryFilter: domain.QueryFilter{
			Query: req.Query,
		},
		SortFilter: domain.SortFilter{
			Direction: req.Direction,
			Column:    req.Column,
		},
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

// @Summary      Enabled a stopword
// @Description  Set IsEnabled field of the stopword to true
// @Tags         Stopwords
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        word path string true "Word"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /stopwords/{word}/enable [post]
func (h *StopwordHandler) Enable(ctx *gin.Context) error {
	return h.withOwnedItemAction(ctx, h.stopwordService.Enable)
}

// @Summary      Disable a stopword
// @Description  Set IsEnabled field of the stopword to false
// @Tags         Stopwords
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        word path string true "Word"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /stopwords/{word}/disable [post]
func (h *StopwordHandler) Disable(ctx *gin.Context) error {
	return h.withOwnedItemAction(ctx, h.stopwordService.Disable)
}

// @Summary      Delete a stopword
// @Description  Deletes a stopword from user profile
// @Tags         Stopwords
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        word path string true "Word"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      404   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /stopwords/{word} [delete]
func (h *StopwordHandler) Delete(ctx *gin.Context) error {
	return h.withOwnedItemAction(ctx, h.stopwordService.Delete)
}

func (h *StopwordHandler) withOwnedItemAction(
	ctx *gin.Context,
	fn func(context.Context, string, string) error,
) error {
	userID := ctx.MustGet("userID").(string)

	var uri stopwordUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	err := fn(ctx.Request.Context(), uri.Word, userID)
	if err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}
