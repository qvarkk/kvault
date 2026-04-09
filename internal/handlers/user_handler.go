package handlers

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/httpx"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	GetByEmail(context.Context, string) (*domain.User, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// TODO: fix
func (h *UserHandler) GetByEmail(ctx *gin.Context) error {
	email := ctx.Query("email")
	if email == "" {
		return httpx.ErrBadRequest
	}

	user, err := h.userService.GetByEmail(ctx.Request.Context(), email)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
	return nil
}
