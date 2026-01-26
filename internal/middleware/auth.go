package middleware

import (
	"context"
	"net/http"
	"qvarkk/kvault/internal/errors"
	"qvarkk/kvault/internal/repo"
	"qvarkk/kvault/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthRequired(userRepo *repo.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		api_key := c.GetHeader("Authorization")

		user, err := userRepo.GetByApiKey(context.Background(), api_key)
		if err != nil {
			_, message := errors.ParseDBError(err)
			logger.Logger.Error("Failed to authorize user by API key", zap.String("message", message))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if user == nil {
			rfc9457_err := errors.FormRFC9457Error(http.StatusUnauthorized, c.FullPath(), "")
			c.AbortWithStatusJSON(http.StatusUnauthorized, rfc9457_err)
			return
		}

		c.Next()
	}
}
