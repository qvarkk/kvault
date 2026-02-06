package handlers

import (
	"context"
	"errors"
	"net/http"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/httpx"
	"qvarkk/kvault/internal/services"

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

func (h *UserHandler) GetByEmail(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		abortWithPublicError(ctx, httpx.ErrBadRequest, "Email query parameter is required.")
		return
	}

	user, err := h.userService.GetByEmail(ctx.Request.Context(), email)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			abortWithPublicError(ctx, httpx.ErrNotFound, "User with this email not found.")
			return
		}
		abortOnDbError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toUserResponse(user))
}
