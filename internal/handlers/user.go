package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/errors"
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
		rfc9457Err := errors.FormRFC9457Error(http.StatusBadRequest, c.FullPath(), "Email query parameter is required.")
		c.JSON(http.StatusBadRequest, rfc9457Err)
		return
	}

	user, err := h.userRepo.GetByEmail(context.Background(), email)
	if err != nil {
		status, message := errors.ParseDBError(err)
		rfc9457Err := errors.FormRFC9457Error(status, c.FullPath(), message)
		c.JSON(status, rfc9457Err)
		return
	}
	if user == nil {
		rfc9457Err := errors.FormRFC9457Error(http.StatusNotFound, c.FullPath(), "User with this email not found.")
		c.JSON(http.StatusNotFound, rfc9457Err)
		return
	}

	c.JSON(http.StatusOK, user)
}
