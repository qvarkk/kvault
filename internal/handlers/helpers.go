package handlers

import (
	"errors"
	"fmt"
	"qvarkk/kvault/internal/httpx"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func abortOnBindError(ctx *gin.Context, err error) {
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
		ctx.Error(fmt.Errorf("couldn't cast bind error to validator.ValidationError: %w", err))
	}

	ctx.Error(publicErr).SetType(gin.ErrorTypePublic)
	ctx.Abort()
}

func abortOnInternalError(ctx *gin.Context, err error) {
	publicErr := &httpx.PublicError{
		Err: httpx.ErrInternalServer,
	}
	ctx.Error(publicErr).SetType(gin.ErrorTypePublic)
	ctx.Error(err)
	ctx.Abort()
}

func abortOnDbError(ctx *gin.Context, err error) {
	publicErr := httpx.DBErrorToPublicError(err)
	ctx.Error(publicErr).SetType(gin.ErrorTypePublic)
	ctx.Abort()
}

func abortUnauthorized(ctx *gin.Context) {
	publicErr := &httpx.PublicError{
		Err: httpx.ErrUnauthorized,
	}
	ctx.Error(publicErr).SetType(gin.ErrorTypePublic)
	ctx.Abort()
}

func abortWithPublicError(ctx *gin.Context, err error, msg string) {
	publicErr := &httpx.PublicError{
		Err:     err,
		Message: msg,
	}
	ctx.Error(publicErr).SetType(gin.ErrorTypePublic)
	ctx.Abort()
}
