package main

import (
	"fmt"
	"log"
	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/migrations"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/internal/repositories"
	"qvarkk/kvault/internal/routes"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/logger"
	"time"

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

	pg, err := postgres.NewPostgres(pgConfig)
	if err != nil {
		logger.Logger.Fatal("Connection to database failed", zap.Error(err))
	}
	defer pg.Close()

	err = migrations.RunMigrations(pg.DB.DB, config.DBDatabase, "file://migrations")
	if err != nil {
		logger.Logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	var (
		userRepo = repositories.NewUserRepo(pg.DB)
	)

	var (
		authService = services.NewAuthService(userRepo)
		userService = services.NewUserService(userRepo)
	)

	services := &routes.Services{
		AuthService:     authService,
		AuthUserService: userService,
		MwUserService:   userService,
		UserService:     userService,
	}

	r := routes.SetupRouter(services)
	r.Run(fmt.Sprintf(":%d", config.ServerPort))
}
