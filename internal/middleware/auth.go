package middleware

import (
	"context"
	"qvarkk/kvault/internal/httpx"
	"qvarkk/kvault/internal/repo"

	"github.com/gin-gonic/gin"
)

func AuthRequired(userRepo *repo.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		api_key := c.GetHeader("Authorization")

		user, err := userRepo.GetByApiKey(context.Background(), api_key)
		if err != nil {
			handlerErr := httpx.DBErrorToPublicError(err)
			c.Error(handlerErr).SetType(gin.ErrorTypePublic)
			c.Abort()
			return
		}
		if user == nil {
			handlerErr := &httpx.PublicError{
				Err: httpx.ErrUnauthorized,
			}
			c.Error(handlerErr).SetType(gin.ErrorTypePublic)
			c.Abort()
			return
		}

		c.Set("authenticatedUser", user)
		c.Next()
	}
}
