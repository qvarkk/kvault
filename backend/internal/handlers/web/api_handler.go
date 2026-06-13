package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIHandler func(c *gin.Context) error

func APIWrap(h APIHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := h(ctx); err != nil {
			_ = ctx.Error(err)
			if ctx.Writer.Status() == http.StatusOK {
				ctx.Status(http.StatusInternalServerError)
			}
			ctx.Abort()
		}
	}
}
