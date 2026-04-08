package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"qvarkk/kvault/internal/domain"

	"github.com/hibiken/asynq"
)

type FileTaskService interface {
	ExtractTextFromFile(context.Context, *domain.File) (string, error)
	UpdateFileStatusByID(context.Context, string, domain.FileStatus) (*domain.File, error)
	UpdateFileTextContentByID(ctx context.Context, fileID string, textContent string) (*domain.File, error)
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

	file, err := h.fileService.UpdateFileStatusByID(ctx, p.FileID, domain.FileStatusProcessing)
	if err != nil {
		return err
	}
	fmt.Printf("File status updated to processing: %s\n", file.Status)

	text, err := h.fileService.ExtractTextFromFile(ctx, file)
	if err != nil {
		return err
	}
	fmt.Printf("Text extracted: %s\n", text[:50])

	file, err = h.fileService.UpdateFileTextContentByID(ctx, p.FileID, text)
	if err != nil {
		return err
	}
	fmt.Printf("Text content updated: %s\n", file.TextContent.String[:50])

	file, err = h.fileService.UpdateFileStatusByID(ctx, p.FileID, domain.FileStatusReady)
	if err != nil {
		return err
	}
	fmt.Printf("File status updated to ready: %s\n", file.Status)

	return nil
}
