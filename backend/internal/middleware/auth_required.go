package middleware

import (
	"context"
	"strings"

	"qvarkk/kvault/internal/domain"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	Authenticate(context.Context, string) (*domain.User, string, error)
}

func AuthRequired(userService UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := strings.TrimSpace(ctx.GetHeader("Authorization"))
		// Accept both a bare key and the conventional "Bearer <key>" form.
		if after, ok := cutBearerPrefix(apiKey); ok {
			apiKey = strings.TrimSpace(after)
		}

		user, apiKeyID, err := userService.Authenticate(ctx.Request.Context(), apiKey)
		if err != nil {
			_ = ctx.Error(err)
			ctx.Abort()
			return
		}

		ctx.Set("userID", user.ID)
		ctx.Set("apiKeyID", apiKeyID)
		ctx.Next()
	}
}

// cutBearerPrefix strips a case-insensitive "Bearer " scheme prefix if present.
func cutBearerPrefix(s string) (string, bool) {
	const prefix = "bearer "
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):], true
	}
	return s, false
}
