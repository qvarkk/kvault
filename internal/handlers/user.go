package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/repo"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo *repo.UserRepo
}

func NewUserHandler(userRepo *repo.UserRepo) *UserHandler {
	return &UserHandler{userRepo: userRepo}
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
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
