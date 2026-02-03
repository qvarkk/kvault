package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repo"
	"qvarkk/kvault/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repo.UserRepo
}

type registerUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type authenticateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(userRepo *repo.UserRepo) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req registerUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(c, err)
		return
	}

	apiKey, err := h.generateApiKey()
	if err != nil {
		abortOnInternalError(c, err)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		abortOnInternalError(c, err)
		return
	}

	user := domain.User{
		Email:    req.Email,
		Password: string(passwordHash),
		APIKey:   apiKey,
	}

	err = h.userRepo.CreateUser(context.Background(), &user)
	if err != nil {
		abortOnDbError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) AuthenticateUser(c *gin.Context) {
	var req authenticateUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		abortOnBindError(c, err)
		return
	}

	user, err := h.userRepo.GetByEmail(context.Background(), req.Email)
	if err != nil {
		abortOnDbError(c, err)
		return
	}
	if user == nil {
		abortUnauthorized(c)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		abortUnauthorized(c)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) GetAuthenticatedUser(c *gin.Context) {
	user := requireAuthenticatedUser(c)
	if user == nil {
		abortUnauthorized(c)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) RefreshApiKey(c *gin.Context) {
	user := requireAuthenticatedUser(c)
	if user == nil {
		abortUnauthorized(c)
		return
	}

	apiKey, err := h.generateApiKey()
	if err != nil {
		abortOnInternalError(c, err)
		return
	}

	err = h.userRepo.UpdateApiKey(context.Background(), user, apiKey)
	if err != nil {
		abortOnDbError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) generateApiKey() (string, error) {
	var apiKey string
	for {
		var err error
		if apiKey, err = utils.GenerateAPIKey(utils.APIKeyLength); err != nil {
			return "", err
		}

		var isKeyUnique bool
		if isKeyUnique, err = h.userRepo.IsAPIKeyUnique(context.Background(), apiKey); err != nil {
			return "", err
		}

		if isKeyUnique {
			break
		}
	}

	return apiKey, nil
}
