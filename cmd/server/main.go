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

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = logger.Init(config.Debug)
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.DB.Username, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Database)
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

	err = migrations.RunMigrations(pg.DB.DB, config.DB.Database, "file://migrations")
	if err != nil {
		logger.Logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	var (
		userRepo = repositories.NewUserRepo(pg.DB)
		itemRepo = repositories.NewItemRepo(pg.DB)
	)

	var (
		authService = services.NewAuthService(userRepo)
		userService = services.NewUserService(userRepo)
		itemService = services.NewItemService(itemRepo)
	)

	services := &routes.Services{
		AuthService:     authService,
		AuthUserService: userService,
		MwUserService:   userService,
		UserService:     userService,
		ItemService:     itemService,
	}

	r := routes.SetupRouter(services)
	r.Run(fmt.Sprintf(":%d", config.Api.Port))
}
