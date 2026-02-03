package middleware

import (
	"net/http"
	"qvarkk/kvault/internal/httpx"
	"qvarkk/kvault/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) <= 0 {
			return
		}

		publicErrs := c.Errors.ByType(gin.ErrorTypePublic)
		if len(publicErrs) > 0 {
			lastErr := publicErrs.Last().Err
			if httpErr, ok := lastErr.(*httpx.PublicError); ok {
				errResponse := httpErr.ToErrorResponse(c.FullPath())
				c.AbortWithStatusJSON(errResponse.Status, errResponse)
			} else {
				errResponse := httpx.NewErrorResponse(
					http.StatusInternalServerError,
					c.FullPath(),
					httpx.ErrInternalServer.Error(),
					nil,
				)
				c.AbortWithStatusJSON(errResponse.Status, errResponse)
			}
		}

		privateErr := c.Errors.ByType(gin.ErrorTypePrivate)
		for _, err := range privateErr {
			// TODO: add more information about request to private errors logged
			// 			 also might be a good idea to bring in something like sentry
			logger.Logger.Error("private error",
				zap.String("path", c.FullPath()),
				zap.String("client_ip", c.ClientIP()),
				zap.String("method", c.Request.Method),
				zap.String("err", err.Err.Error()),
			)
		}
	}
}
