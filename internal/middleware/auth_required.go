package middleware

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/httpx"
	"qvarkk/kvault/internal/services"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	GetByApiKey(context.Context, string) (*domain.User, error)
}

func AuthRequired(userService UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		api_key := ctx.GetHeader("Authorization")

		user, err := userService.GetByApiKey(ctx.Request.Context(), api_key)
		if err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				handlerErr := &httpx.PublicError{
					Err: httpx.ErrUnauthorized,
				}
				ctx.Error(handlerErr).SetType(gin.ErrorTypePublic)
				ctx.Abort()
				return
			}
			handlerErr := httpx.DBErrorToPublicError(err)
			ctx.Error(handlerErr).SetType(gin.ErrorTypePublic)
			ctx.Abort()
			return
		}

		ctx.Set("userID", user.ID)
		ctx.Next()
	}
}
