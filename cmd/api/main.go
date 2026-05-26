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
	"time"

	"go.uber.org/zap"
)

// @title           KVault API
// @version         1.0
// @description     REST API for managing and searching notes and documents
// @host            172.21.37.79:6767
// @BasePath        /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
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
		zap.L().Fatal("Connection to database failed", zap.Error(err))
	}
	defer pg.Close()

	redisConnConfig := redis.ConnConfig{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Username: config.Redis.User,
		Password: config.Redis.Password,
	}
	queueConfig := redisConnConfig
	queueConfig.DB = config.Redis.QueueDb

	enqueuer, err := redis.NewAsynqEnqueuer(queueConfig)
	if err != nil {
		zap.L().Fatal("Asynq connection to Redis failed", zap.Error(err))
	}

	cacheConfig := redis.CacheConfig{
		IsEnabled:    config.Cache.Enabled,
		ItemsTtl:     config.Cache.ItemsTtl,
		FilesTtl:     config.Cache.FilesTtl,
		TagsTtl:      config.Cache.TagsTtl,
		StopwordsTtl: config.Cache.StopwordsTtl,
	}

	var cacheStore services.CacheStore
	if config.Cache.Enabled {
		storeConfig := redisConnConfig
		storeConfig.DB = config.Redis.CacheDb
		redisClient, err := redis.NewRedisStore(storeConfig, cacheConfig)
		if err != nil {
			zap.L().Fatal("Cache connection to Redis failed", zap.Error(err))
		}
		cacheStore = redisClient
	} else {
		cacheStore = redis.NewNoopCache()
	}

	aws, err := aws.NewAwsStorage(config.Aws, config.Api.CorsOrigins)
	if err != nil {
		zap.L().Fatal("Connection to AWS failed", zap.Error(err))
	}

	var (
		userRepo     = repositories.NewUserRepo(pg.DB)
		itemRepo     = repositories.NewItemRepo(pg.DB)
		fileRepo     = repositories.NewFileRepo(pg.DB)
		stopwordRepo = repositories.NewStopwordRepo(pg.DB)
		tagRepo      = repositories.NewTagRepo(pg.DB)
		transactor   = repositories.NewTransactor(pg.DB)
	)

	var (
		authService     = services.NewAuthService(userRepo)
		userService     = services.NewUserService(userRepo)
		itemService     = services.NewItemService(itemRepo, tagRepo, transactor, cacheStore, enqueuer, cacheConfig.ItemsTtl)
		fileService     = services.NewFileService(fileRepo, transactor, enqueuer, aws, cacheStore, cacheConfig.FilesTtl)
		stopwordService = services.NewStopwordService(stopwordRepo, transactor, cacheStore, cacheConfig.StopwordsTtl)
		tagService      = services.NewTagService(tagRepo, stopwordRepo, itemRepo, transactor, cacheStore, cacheConfig.TagsTtl)
	)

	hs := &routes.HandlerServices{
		Auth:     authService,
		AuthUser: userService,
		User:     userService,
		Item:     itemService,
		File:     fileService,
		Stopword: stopwordService,
		Tag:      tagService,
	}

	ms := &routes.MiddlewareServices{
		User: userService,
	}

	r := routes.SetupRouter(hs, ms, config.Api.CorsOrigins)
	r.Run(fmt.Sprintf(":%d", config.Api.Port))
}
