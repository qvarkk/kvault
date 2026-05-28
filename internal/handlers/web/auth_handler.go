package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	GenerateApiKey(context.Context) (string, error)
	RegisterNewUser(ctx context.Context, username string, password string) (*domain.User, error)
	VerifyCredentials(ctx context.Context, username string, password string) (*domain.User, error)
	RotateApiKey(ctx context.Context, userID string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	VerifyPassword(ctx context.Context, userID, password string) error
	DeleteAccount(ctx context.Context, userID string) error
}

type AuthUserService interface {
	GetByID(context.Context, string) (*domain.User, error)
}

type AuthFileService interface {
	DeleteAllByUserID(ctx context.Context, userID string) error
}

type AuthHandler struct {
	authService AuthService
	userService AuthUserService
	fileService AuthFileService
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type deleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

type registerUserRequest struct {
	Username string `json:"username" binding:"required,min=3" example:"john_doe"`
	Password string `json:"password" binding:"required,min=8" example:"#strongPwd?123."`
}

type authenticateUserRequest struct {
	Username string `json:"username" binding:"required,min=3" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"#strongPwd?123."`
}

func NewAuthHandler(authService AuthService, userService AuthUserService, fileService AuthFileService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		fileService: fileService,
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
func (h *AuthHandler) RegisterUser(ctx *gin.Context) error {
	var req registerUserRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	user, err := h.authService.RegisterNewUser(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusCreated, toUserResponseWithApiKey(user))
	return nil
}

// @Summary      User authentication
// @Description  Verifies user credentials
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body body authenticateUserRequest true "User credentials"
// @Success      200   {object}  UserResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) AuthenticateUser(ctx *gin.Context) error {
	var req authenticateUserRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	user, err := h.authService.VerifyCredentials(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toUserResponseWithApiKey(user))
	return nil
}

// @Summary      Get user data
// @Description  Returns authenticated user
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200   {object}  UserResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/me [get]
func (h *AuthHandler) GetAuthenticatedUser(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	user, err := h.userService.GetByID(ctx.Request.Context(), userID)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toUserResponseWithApiKey(user))
	return nil
}

// @Summary      Refresh API key
// @Description  Refreshes authenticated user's API key
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200   {object}  UserResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RotateApiKey(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	user, err := h.authService.RotateApiKey(ctx, userID)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toUserResponseWithApiKey(user))
	return nil
}

// @Summary      Change password
// @Description  Verifies old password and replaces it with new password
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Accept       json
// @Param        body body changePasswordRequest true "Passwords"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/me/password [patch]
func (h *AuthHandler) ChangePassword(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req changePasswordRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	if err := h.authService.ChangePassword(ctx.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Delete account
// @Description  Verifies password and permanently deletes the authenticated user and all their data
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Accept       json
// @Param        body body deleteAccountRequest true "Password confirmation"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/me/delete [post]
func (h *AuthHandler) DeleteAccount(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var req deleteAccountRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	if err := h.authService.VerifyPassword(ctx.Request.Context(), userID, req.Password); err != nil {
		return err
	}

	_ = h.fileService.DeleteAllByUserID(ctx.Request.Context(), userID)

	if err := h.authService.DeleteAccount(ctx.Request.Context(), userID); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}
