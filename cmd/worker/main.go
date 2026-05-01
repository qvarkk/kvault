package main

import (
	"fmt"
	"log"
	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/aws"
	"qvarkk/kvault/internal/handlers/worker"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/internal/repositories"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/internal/tasks"
	"qvarkk/kvault/logger"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = logger.Init("worker", config.Debug)
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

	aws, err := aws.NewAws(config.Aws)
	if err != nil {
		logger.Logger.Fatal("Connection to AWS failed", zap.Error(err))
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
			Username: config.Redis.User,
			Password: config.Redis.Password,
			DB:       0,
		},
		asynq.Config{Concurrency: config.Worker.ConcurrentTasks},
	)

	fileRepo := repositories.NewFileRepo(pg.DB)
	transactor := repositories.NewTransactor(pg.DB)
	fileService := services.NewFileTaskService(fileRepo, transactor, aws)
	fileTaskHandler := worker.NewFileTaskHandler(fileService)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypePdfProcess, fileTaskHandler.HandlePdfProcessTask)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
