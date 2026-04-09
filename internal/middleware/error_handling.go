package middleware

import (
	"errors"
	"qvarkk/kvault/internal/httpx"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) <= 0 {
			return
		}

		err := c.Errors.Last().Err

		var publicErr *httpx.PublicError
		var serviceErr *services.ServiceError
		var validationErr validator.ValidationErrors

		switch {
		case errors.As(err, &publicErr):

		case errors.As(err, &serviceErr):
			publicErr = httpx.MapErrorToPublic(err)
		case errors.As(err, &validationErr):
			publicErr = &httpx.PublicError{
				Err:              httpx.ErrUnprocessableEntity,
				ValidationErrors: validationErr,
			}
		default:
			publicErr = &httpx.PublicError{
				Err: httpx.ErrInternalServer,
			}
		}

		for _, ginErr := range c.Errors {
			logger.Logger.Error("request error",
				zap.String("path", c.FullPath()),
				zap.String("method", c.Request.Method),
				zap.Error(ginErr.Err),
			)
		}

		errResponse := publicErr.ToErrorResponse(c.FullPath())
		c.AbortWithStatusJSON(errResponse.Status, errResponse)
	}
}
