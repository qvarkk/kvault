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
	RotateApiKey(ctx context.Context, userID string) (*domain.User, error)
}

type AuthUserService interface {
	GetByID(context.Context, string) (*domain.User, error)
}

type AuthHandler struct {
	authService AuthService
	userService AuthUserService
}

type registerUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"example@mail.com"`
	Password string `json:"password" binding:"required,min=8" example:"#strongPwd?123."`
}

type authenticateUserRequest struct {
	Email    string `json:"email" binding:"required,email" example:"example@mail.com"`
	Password string `json:"password" binding:"required" example:"#strongPwd?123."`
}

func NewAuthHandler(authService AuthService, userService AuthUserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// @Summary      User registration
// @Description  Creates a user record in database with given credentials
// @Description  and returns user's information
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body registerUserRequest true "User credentials"
// @Success      201   {object}  UserResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) RegisterUser(ctx *gin.Context) {
	var req registerUserRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(ctx, err)
		return
	}

	user, err := h.authService.RegisterNewUser(ctx.Request.Context(), req.Email, req.Password)
	if errors.Is(err, services.ErrUserNotCreated) {
		abortOnDbError(ctx, err)
		return
	} else if errors.Is(err, services.ErrInternal) {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toUserResponseWithApiKey(user))
}

// @Summary      User authentication
// @Description  Verifies user credentials
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body authenticateUserRequest true "User credentials"
// @Success      200   {object}  UserResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/login [post]
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

	ctx.JSON(http.StatusOK, toUserResponseWithApiKey(user))
}

// @Summary      Get user data
// @Description  Returns authenticated user
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200   {object}  UserResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/me [get]
func (h *AuthHandler) GetAuthenticatedUser(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	user, err := h.userService.GetByID(ctx.Request.Context(), userID)
	if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toUserResponseWithApiKey(user))
}

// @Summary      Refresh API key
// @Description  Refreshes authenticated user's API key
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200   {object}  UserResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RotateApiKey(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	user, err := h.authService.RotateApiKey(ctx, userID)
	if err != nil {
		abortOnInternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toUserResponseWithApiKey(user))
}
