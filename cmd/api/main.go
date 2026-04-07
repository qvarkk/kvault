package main

import (
	"fmt"
	"log"
	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/aws"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/internal/redis"
	"qvarkk/kvault/internal/repositories"
	"qvarkk/kvault/internal/routes"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/logger"
	"qvarkk/kvault/migrations"
	"time"

	"go.uber.org/zap"
)

// @title           KVault API
// @version         1.0
// @description     REST API for managing and searching notes and documents
// @host            localhost:6767
// @BasePath        /api/v1
func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = logger.Init("api", config.Debug)
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

	err = migrations.RunMigrations(pg.DB.DB, config.DB.Database)
	if err != nil {
		logger.Logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	redisConfig := redis.Config{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Username: config.Redis.User,
		Password: config.Redis.Password,
		DB:       0,
	}

	redis, err := redis.NewRedis(redisConfig)
	if err != nil {
		logger.Logger.Fatal("Connection to Redis failed", zap.Error(err))
	}

	aws, err := aws.NewAws(config.Aws)
	if err != nil {
		logger.Logger.Fatal("Failed loading config", zap.Error(err))
	}

	var (
		userRepo = repositories.NewUserRepo(pg.DB)
		itemRepo = repositories.NewItemRepo(pg.DB)
		fileRepo = repositories.NewFileRepo(pg.DB)
	)

	var (
		authService = services.NewAuthService(userRepo)
		userService = services.NewUserService(userRepo)
		itemService = services.NewItemService(itemRepo)
		fileService = services.NewFileService(fileRepo, redis, aws)
	)

	services := &routes.Services{
		AuthService:     authService,
		AuthUserService: userService,
		MwUserService:   userService,
		UserService:     userService,
		ItemService:     itemService,
		FileService:     fileService,
	}

	r := routes.SetupRouter(services)
	r.Run(fmt.Sprintf(":%d", config.Api.Port))
}
