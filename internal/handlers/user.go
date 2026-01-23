package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/repo"
	"qvarkk/kvault/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo *repo.UserRepo
}

func NewUserHandler(userRepo *repo.UserRepo) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

type createUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type authenticateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		errors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors})
		return
	}

	var APIKey string
	for {
		var err error
		if APIKey, err = utils.GenerateAPIKey(utils.APIKeyLength); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var isKeyUnique bool
		if isKeyUnique, err = h.userRepo.IsAPIKeyUnique(context.Background(), APIKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if isKeyUnique {
			break
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user := domain.User{
		Email:     req.Email,
		Password:  string(passwordHash),
		APIKey:    APIKey,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = h.userRepo.CreateUser(context.Background(), &user)
	if err != nil {
		status, message := utils.ParseDBError(err)
		c.JSON(status, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) AuthenticateUser(c *gin.Context) {
	var req authenticateUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		errors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors})
		return
	}

	user, err := h.userRepo.AuthenticateUser(c, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong credentials"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter is required"})
		return
	}

	user, err := h.userRepo.GetByEmail(context.Background(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
