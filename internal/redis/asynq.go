package redis

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/tasks"

	"github.com/hibiken/asynq"
)

type AsynqEnqueuer struct {
	client *asynq.Client
}

func NewAsynqEnqueuer(config Config) (*AsynqEnqueuer, error) {
	asynqOpt := asynq.RedisClientOpt{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	}

	client := asynq.NewClient(asynqOpt)

	if err := client.Ping(); err != nil {
		cErr := client.Close()
		return nil, errors.Join(err, cErr)
	}

	return &AsynqEnqueuer{
		client: client,
	}, nil
}

func (e *AsynqEnqueuer) EnqueuePdfProcess(ctx context.Context, userID, fileID string) error {
	payload := tasks.PdfProcessPayload{
		UserID: userID,
		FileID: fileID,
	}

	task, err := tasks.NewPdfProcessTask(payload)
	if err != nil {
		return err
	}

	_, err = e.client.EnqueueContext(ctx, task)
	return err
}
