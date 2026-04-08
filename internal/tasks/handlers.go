package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type FileTaskService interface {
	ExtractTextFromS3(context.Context, string) (string, error)
}

type FileTaskHandler struct {
	fileService FileTaskService
}

func NewFileTaskHandler(fileService FileTaskService) *FileTaskHandler {
	return &FileTaskHandler{
		fileService: fileService,
	}
}

func (h *FileTaskHandler) HandlePdfProcessTask(ctx context.Context, t *asynq.Task) error {
	var p PdfProcessPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	text, err := h.fileService.ExtractTextFromS3(ctx, p.FileID)
	if err != nil {
		return err
	}

	fmt.Printf("Extracted text:\n%s\n", text)

	return nil
}
