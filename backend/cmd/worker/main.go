package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/handlers/worker"
	"qvarkk/kvault/internal/postgres"
	"qvarkk/kvault/internal/repositories"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/internal/storage"
	"qvarkk/kvault/internal/tasks"
	"qvarkk/kvault/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = logger.Init("worker", cfg.Debug)
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Database)
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
	defer func() {
		if err := pg.Close(); err != nil {
			zap.L().Error("failed to close database connection", zap.Error(err))
		}
	}()

	localStorage, err := storage.NewLocalStorage(&cfg.Storage)
	if err != nil {
		zap.L().Error("local storage init failed", zap.Error(err))
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Username: cfg.Redis.User,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.QueueDb,
		},
		asynq.Config{
			Concurrency: cfg.Worker.ConcurrentTasks,
			ErrorHandler: asynq.ErrorHandlerFunc(func(_ context.Context, task *asynq.Task, err error) {
				zap.L().Error("asynq task failed",
					zap.String("type", task.Type()),
					zap.ByteString("payload", task.Payload()),
					zap.Error(err),
				)
			}),
		},
	)

	fileRepo := repositories.NewFileRepo(pg.DB)
	transactor := repositories.NewTransactor(pg.DB)
	fileService := services.NewFileTaskService(fileRepo, transactor, localStorage)
	fileTaskHandler := worker.NewFileTaskHandler(fileService)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypePdfProcess, fileTaskHandler.HandlePdfProcessTask)

	if err := srv.Run(mux); err != nil {
		zap.L().Error("asynq worker failed",
			zap.Error(err),
		)
	}
}
