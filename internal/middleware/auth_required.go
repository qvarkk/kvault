package middleware

import (
	"context"
	"qvarkk/kvault/internal/domain"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	Authenticate(context.Context, string) (*domain.User, error)
}

func AuthRequired(userService UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		api_key := ctx.GetHeader("Authorization")

		user, err := userService.Authenticate(ctx.Request.Context(), api_key)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		ctx.Set("userID", user.ID)
		ctx.Next()
	}
}
