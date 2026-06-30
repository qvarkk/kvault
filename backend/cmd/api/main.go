package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/internal/redis"
	"qvarkk/kvault/internal/repositories"
	"qvarkk/kvault/internal/routes"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/internal/storage"
	"qvarkk/kvault/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title           KVault API
// @version         1.0
// @description     REST API for managing and searching notes and documents
// @BasePath        /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	err = logger.Init("api", cfg.Debug)
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}

	if err := run(cfg); err != nil {
		zap.L().Error("application terminated unexpectedly", zap.Error(err))
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)

	pgConfig := postgres.Config{
		DSN:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
	}

	pg, err := postgres.NewPostgres(pgConfig)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer func() {
		if err := pg.Close(); err != nil {
			zap.L().Error("failed to close database connection", zap.Error(err))
		}
	}()

	redisConnConfig := redis.ConnConfig{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
	}
	queueConfig := redisConnConfig
	queueConfig.DB = cfg.Redis.QueueDb

	enqueuer, err := redis.NewAsynqEnqueuer(queueConfig, cfg.Worker.MaxRetries, cfg.Worker.RetryTimeout)
	if err != nil {
		return fmt.Errorf("asynq redis connection failed: %w", err)
	}

	cacheConfig := redis.CacheConfig{
		IsEnabled:    cfg.Cache.Enabled,
		ItemsTtl:     cfg.Cache.ItemsTtl,
		FilesTtl:     cfg.Cache.FilesTtl,
		TagsTtl:      cfg.Cache.TagsTtl,
		StopwordsTtl: cfg.Cache.StopwordsTtl,
	}

	var cacheStore services.CacheStore
	if cfg.Cache.Enabled {
		storeConfig := redisConnConfig
		storeConfig.DB = cfg.Redis.CacheDb
		redisClient, err := redis.NewRedisStore(storeConfig, cacheConfig)
		if err != nil {
			return fmt.Errorf("cache redis connection failed: %w", err)
		}
		cacheStore = redisClient
	} else {
		cacheStore = redis.NewNoopCache()
	}

	localStorage, err := storage.NewLocalStorage(&cfg.Storage)
	if err != nil {
		return fmt.Errorf("local storage init failed: %w", err)
	}

	var (
		userRepo   = repositories.NewUserRepo(pg.DB)
		apiKeyRepo = repositories.NewApiKeyRepo(pg.DB)
		itemRepo   = repositories.NewItemRepo(pg.DB)
		fileRepo   = repositories.NewFileRepo(pg.DB)
		tagRepo    = repositories.NewTagRepo(pg.DB)
		transactor = repositories.NewTransactor(pg.DB)
	)

	var (
		authService = services.NewAuthService(userRepo, apiKeyRepo, cfg.Auth.ApiKeyTtl)
		userService = services.NewUserService(userRepo, apiKeyRepo, cfg.Auth.ApiKeyTtl)
		itemService = services.NewItemService(itemRepo, tagRepo, transactor, cacheStore, enqueuer, cacheConfig.ItemsTtl)
		fileService = services.NewFileService(fileRepo, transactor, enqueuer, localStorage, cacheStore, cacheConfig.FilesTtl)
		tagService  = services.NewTagService(tagRepo, itemRepo, transactor, cacheStore, cacheConfig.TagsTtl)
	)

	hs := &routes.HandlerServices{
		Auth:     authService,
		AuthUser: userService,
		Item:     itemService,
		File:     fileService,
		Tag:      tagService,
	}

	ms := &routes.MiddlewareServices{
		User: userService,
	}

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := routes.SetupRouter(hs, ms, cfg.Api.CorsOrigins)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Api.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				zap.L().Info("server closed under request")
			} else {
				zap.L().Error("http server failed", zap.Error(err))
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("receive interrupt signal")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("failed to shutdown http server", zap.Error(err))
	}

	zap.L().Info("server exiting")
	return nil
}
