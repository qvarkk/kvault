package tasks

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

func HandleFileUploadTask(ctx context.Context, t *asynq.Task) error {
	var p FileUploadPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf(" [*] Uploaded file as User %d", p.UserID)
	return nil
}
