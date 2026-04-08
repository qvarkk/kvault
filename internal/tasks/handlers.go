package tasks

import (
	"context"
	"encoding/json"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
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

func (h *FileTaskHandler) HandlePdfProcessTask(ctx context.Context, t *asynq.Task) (err error) {
	var p PdfProcessPayload
	if err = json.Unmarshal(t.Payload(), &p); err != nil {
		logger.Logger.Error("Failed to parse task payload", zap.Error(err), zap.String("file_id", p.FileID))
		return err
	}

	defer func() {
		if err != nil && p.FileID != "" {
			_, updateErr := h.fileService.UpdateFileStatusByID(context.Background(), p.FileID, domain.FileStatusError)
			if updateErr != nil {
				logger.Logger.Error(
					"Failed to update file status to error",
					zap.Error(updateErr),
					zap.String("file_id", p.FileID),
				)
			}
		}
	}()

	var file *domain.File
	file, err = h.fileService.UpdateFileStatusByID(ctx, p.FileID, domain.FileStatusProcessing)
	if err != nil {
		return err
	}

	var text string
	text, err = h.fileService.ExtractTextFromFile(ctx, file)
	if err != nil {
		return err
	}

	file, err = h.fileService.UpdateFileTextContentByID(ctx, p.FileID, text)
	if err != nil {
		return err
	}

	file, err = h.fileService.UpdateFileStatusByID(ctx, p.FileID, domain.FileStatusReady)
	if err != nil {
		return err
	}

	return nil
}
