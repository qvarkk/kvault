package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type FileTaskService interface {
	GetPdfFileFromS3(context.Context, string) (*bytes.Buffer, error)
	ConvertPDFToPlainText(context.Context, *bytes.Reader) (string, error)
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

	file, err := h.fileService.GetPdfFileFromS3(ctx, p.FileID)
	if err != nil {
		return err
	}

	text, err := h.fileService.ConvertPDFToPlainText(ctx, bytes.NewReader(file.Bytes()))
	if err != nil {
		return err
	}

	fmt.Printf("Extracted text:\n%s\n", text)

	return nil
}
