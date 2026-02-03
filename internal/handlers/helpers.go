package handlers

import (
	"errors"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/httpx"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func abortOnBindError(c *gin.Context, err error) {
	var publicErr *httpx.PublicError

	var vErr validator.ValidationErrors
	if errors.As(err, &vErr) {
		publicErr = &httpx.PublicError{
			Err:              httpx.ErrUnprocessableEntity,
			ValidationErrors: vErr,
		}
	} else {
		publicErr = &httpx.PublicError{
			Err: httpx.ErrInternalServer,
		}
		c.Error(fmt.Errorf("couldn't cast bind error to validator.ValidationError: %w", err))
	}

	c.Error(publicErr).SetType(gin.ErrorTypePublic)
	c.Abort()
}

func abortOnInternalError(c *gin.Context, err error) {
	publicErr := &httpx.PublicError{
		Err: httpx.ErrInternalServer,
	}
	c.Error(publicErr).SetType(gin.ErrorTypePublic)
	c.Error(err)
	c.Abort()
}

func abortOnDbError(c *gin.Context, err error) {
	publicErr := httpx.DBErrorToPublicError(err)
	c.Error(publicErr).SetType(gin.ErrorTypePublic)
	c.Abort()
}

func abortUnauthorized(c *gin.Context) {
	publicErr := &httpx.PublicError{
		Err: httpx.ErrUnauthorized,
	}
	c.Error(publicErr).SetType(gin.ErrorTypePublic)
	c.Abort()
}

func abortWithPublicError(c *gin.Context, err error, msg string) {
	publicErr := &httpx.PublicError{
		Err:     err,
		Message: msg,
	}
	c.Error(publicErr).SetType(gin.ErrorTypePublic)
	c.Abort()
}

func requireAuthenticatedUser(c *gin.Context) *domain.User {
	userInterface, exists := c.Get("authenticatedUser")
	if !exists || userInterface == nil {
		return nil
	}

	user, ok := userInterface.(*domain.User)
	if !ok {
		c.Error(fmt.Errorf("couldn't cast user to *domain.User, got %T", user))
		return nil
	}

	return user
}
