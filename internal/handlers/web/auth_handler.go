package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthService interface {
	RegisterNewUser(ctx context.Context, username string, password string, label string) (*domain.User, string, error)
	VerifyCredentials(ctx context.Context, username string, password string, label string) (*domain.User, string, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	VerifyPassword(ctx context.Context, userID, password string) error
	DeleteAccount(ctx context.Context, userID string) error
	ListKeys(ctx context.Context, userID string) ([]domain.ApiKey, error)
	RenameKey(ctx context.Context, userID, keyID, label string) error
	DeleteKey(ctx context.Context, userID, keyID string) error
	Logout(ctx context.Context, userID, keyID string) error
	LogoutOthers(ctx context.Context, userID, keyID string) error
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

	user, apiKey, err := h.authService.RegisterNewUser(ctx.Request.Context(), req.Username, req.Password, deviceLabel(ctx))
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusCreated, toUserResponseWithApiKey(user, apiKey))
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

	user, apiKey, err := h.authService.VerifyCredentials(ctx.Request.Context(), req.Username, req.Password, deviceLabel(ctx))
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toUserResponseWithApiKey(user, apiKey))
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

	ctx.JSON(http.StatusOK, toUserResponse(user))
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

	// Best-effort S3 cleanup before the cascading DB delete. Individual object
	// failures are already logged inside the service; log a top-level failure too
	// so orphaned objects are at least traceable.
	if err := h.fileService.DeleteAllByUserID(ctx.Request.Context(), userID); err != nil {
		zap.L().Warn("failed to delete user's files from storage during account deletion",
			zap.String("user_id", userID), zap.Error(err))
	}

	if err := h.authService.DeleteAccount(ctx.Request.Context(), userID); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

type keyIDUri struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type renameKeyRequest struct {
	Label string `json:"label" binding:"required,max=64"`
}

// @Summary      List API keys
// @Description  Lists the authenticated user's active API keys (one per device)
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200   {array}   ApiKeyResponse
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/keys [get]
func (h *AuthHandler) ListKeys(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)
	currentKeyID := ctx.MustGet("apiKeyID").(string)

	keys, err := h.authService.ListKeys(ctx.Request.Context(), userID)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toApiKeyResponses(keys, currentKeyID))
	return nil
}

// @Summary      Rename an API key
// @Description  Updates the label of one of the authenticated user's API keys
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Accept       json
// @Param        id   path string true "API key ID"
// @Param        body body renameKeyRequest true "New label"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/keys/{id} [patch]
func (h *AuthHandler) RenameKey(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri keyIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	var req renameKeyRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	if err := h.authService.RenameKey(ctx.Request.Context(), userID, uri.ID, req.Label); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Delete an API key
// @Description  Revokes one of the authenticated user's API keys, logging out that device
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Param        id path string true "API key ID"
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      422   {object}  httpx.ErrorResponse "Validation Error"
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/keys/{id} [delete]
func (h *AuthHandler) DeleteKey(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)

	var uri keyIDUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		return err
	}

	if err := h.authService.DeleteKey(ctx.Request.Context(), userID, uri.ID); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Log out
// @Description  Revokes the API key used for the current request
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)
	keyID := ctx.MustGet("apiKeyID").(string)

	if err := h.authService.Logout(ctx.Request.Context(), userID, keyID); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}

// @Summary      Log out other devices
// @Description  Revokes every API key except the one used for the current request
// @Tags         Authentication
// @Security     ApiKeyAuth
// @Success      204
// @Failure      401   {object}  httpx.ErrorResponse
// @Failure      500   {object}  httpx.ErrorResponse
// @Router       /auth/logout-others [post]
func (h *AuthHandler) LogoutOthers(ctx *gin.Context) error {
	userID := ctx.MustGet("userID").(string)
	keyID := ctx.MustGet("apiKeyID").(string)

	if err := h.authService.LogoutOthers(ctx.Request.Context(), userID, keyID); err != nil {
		return err
	}

	ctx.Status(http.StatusNoContent)
	return nil
}
