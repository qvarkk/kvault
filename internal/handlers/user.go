package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/httpx"
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
		abortWithPublicError(c, httpx.ErrBadRequest, "Email query parameter is required.")
		return
	}

	user, err := h.userRepo.GetByEmail(context.Background(), email)
	if err != nil {
		abortOnDbError(c, err)
		return
	}
	if user == nil {
		abortWithPublicError(c, httpx.ErrNotFound, "User with this email not found.")
		return
	}

	c.JSON(http.StatusOK, user)
}
