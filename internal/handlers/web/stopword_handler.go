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
	List(context.Context, domain.ListStopwordParams) ([]domain.Stopword, error)
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
