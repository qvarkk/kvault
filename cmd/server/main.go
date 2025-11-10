package main

import (
	"fmt"
	"log"
	"net/http"
	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/logger"
	"time"

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

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.DBUsername, config.DBPassword, config.DBHost, config.DBPort, config.DBDatabase)
	pgConfig := postgres.Config{
		DSN:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
	}

	pg, err := postgres.InitPostgres(pgConfig)
	if err != nil {
		logger.Logger.Fatal("Connection to database failed", zap.Error(err))
	}
	defer pg.Close()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()
}
