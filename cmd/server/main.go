package main

import (
	"log"
	"net/http"
	"qvarkk/kvault/config"
	"qvarkk/kvault/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var DEBUG bool = true

func main() {
	err := logger.Init(DEBUG)
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	config, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal("Failed to load config", zap.Error(err))
	}
	_ = config

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()
}