package web

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/httpx"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	GetByUsername(context.Context, string) (*domain.User, error)
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// TODO: fix
func (h *UserHandler) GetByUsername(ctx *gin.Context) error {
	username := ctx.Query("username")
	if username == "" {
		return httpx.ErrBadRequest
	}

	user, err := h.userService.GetByUsername(ctx.Request.Context(), username)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
	return nil
}
