package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

func HandlePdfProcessTask(ctx context.Context, t *asynq.Task) error {
	var p PdfProcessPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	fmt.Printf("[*] Processing PDF\n\tUser: %s\n\tFile: %s\n", p.UserID, p.FileID)
	return nil
}
