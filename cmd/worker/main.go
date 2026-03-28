package main

import (
	"fmt"
	"log"
	"qvarkk/kvault/config"
	"qvarkk/kvault/internal/tasks"
	"qvarkk/kvault/logger"

	"github.com/hibiken/asynq"
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

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
			Username: config.Redis.User,
			Password: config.Redis.Password,
			DB:       0,
		},
		asynq.Config{Concurrency: config.Worker.ConcurrentTasks},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeFileUpload, tasks.HandleFileUploadTask)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
