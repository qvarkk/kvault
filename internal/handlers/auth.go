package handlers

import (
	"context"
	"errors"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	GenerateApiKey(context.Context) (string, error)
	RegisterNewUser(ctx context.Context, email string, password string) (*domain.User, error)
	VerifyCredentials(ctx context.Context, email string, password string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	RotateApiKey(ctx context.Context, userID string) (*domain.User, error)
}

type AuthHandler struct {
	authService AuthService
}

type registerUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type authenticateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterUser(ctx *gin.Context) {
	var req registerUserRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	user, err := h.authService.RegisterNewUser(ctx.Request.Context(), req.Email, req.Password)
	if errors.Is(err, services.ErrDatabase) {
		abortOnDbError(ctx, err)
		return
	} else if errors.Is(err, services.ErrInternal) {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *AuthHandler) AuthenticateUser(ctx *gin.Context) {
	var req authenticateUserRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	user, err := h.authService.VerifyCredentials(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		abortUnauthorized(ctx)
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) GetAuthenticatedUser(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	user, err := h.authService.GetUserByID(ctx.Request.Context(), userID)
	if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) RotateApiKey(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	user, err := h.authService.RotateApiKey(ctx, userID)
	if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}
