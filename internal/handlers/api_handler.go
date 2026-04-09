package handlers

import (
	"github.com/gin-gonic/gin"
)

type APIHandler func(c *gin.Context) error

func APIWrap(h APIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			c.Error(err)
			c.Abort()
		}
	}
}
