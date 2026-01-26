package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/errors"
	"qvarkk/kvault/internal/repo"
	"qvarkk/kvault/internal/utils"
	"qvarkk/kvault/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repo.UserRepo
}

func NewAuthHandler(userRepo *repo.UserRepo) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

type registerUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type authenticateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req registerUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		rfc9457Err := errors.FormRFC9457Error(http.StatusBadRequest, c.FullPath(), "")
		rfc9457Err.Validation = errors.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, rfc9457Err)
		return
	}

	var apiKey string
	for {
		var err error
		if apiKey, err = utils.GenerateAPIKey(utils.APIKeyLength); err != nil {
			rfc9457Err := errors.FormRFC9457Error(http.StatusInternalServerError, c.FullPath(), "")
			logger.Logger.Error("Failed to generate API key", zap.Error(err))
			c.JSON(http.StatusInternalServerError, rfc9457Err)
			return
		}

		var isKeyUnique bool
		if isKeyUnique, err = h.userRepo.IsAPIKeyUnique(context.Background(), apiKey); err != nil {
			rfc9457Err := errors.FormRFC9457Error(http.StatusInternalServerError, c.FullPath(), "")
			logger.Logger.Error("Failed to check if generated API key is unique", zap.Error(err), zap.String("key", apiKey))
			c.JSON(http.StatusInternalServerError, rfc9457Err)
			return
		}

		if isKeyUnique {
			break
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		rfc9457Err := errors.FormRFC9457Error(http.StatusInternalServerError, c.FullPath(), "")
		logger.Logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, rfc9457Err)
		return
	}

	user := domain.User{
		Email:    req.Email,
		Password: string(passwordHash),
		APIKey:   apiKey,
	}

	err = h.userRepo.CreateUser(context.Background(), &user)
	if err != nil {
		status, message := errors.ParseDBError(err)
		rfc9457Err := errors.FormRFC9457Error(status, c.FullPath(), message)
		c.JSON(status, rfc9457Err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) AuthenticateUser(c *gin.Context) {
	var req authenticateUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		rfc9457Err := errors.FormRFC9457Error(http.StatusBadRequest, c.FullPath(), "")
		rfc9457Err.Validation = errors.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, rfc9457Err)
		return
	}

	user, err := h.userRepo.GetByEmail(c, req.Email)
	if err != nil {
		status, message := errors.ParseDBError(err)
		rfc9457Err := errors.FormRFC9457Error(status, c.FullPath(), message)
		c.JSON(status, rfc9457Err)
		return
	}
	if user == nil {
		rfc9457Err := errors.FormRFC9457Error(http.StatusUnauthorized, c.FullPath(), "")
		c.JSON(http.StatusUnauthorized, rfc9457Err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		rfc9457Err := errors.FormRFC9457Error(http.StatusUnauthorized, c.FullPath(), "")
		c.JSON(http.StatusUnauthorized, rfc9457Err)
		return
	}

	c.JSON(http.StatusOK, user)
}
