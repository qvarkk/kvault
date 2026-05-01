package worker

import (
	"context"
	"encoding/json"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/services"
	"qvarkk/kvault/internal/tasks"
	"qvarkk/kvault/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type FileTaskService interface {
	ExtractTextFromFile(context.Context, *domain.File) (string, error)
	UpdateFile(context.Context, services.UpdateFileInput) (*domain.File, error)
}

type FileTaskHandler struct {
	fileService FileTaskService
}

// helper to get poiners from any values (including const)
func Ptr[T any](v T) *T { return &v }

func NewFileTaskHandler(fileService FileTaskService) *FileTaskHandler {
	return &FileTaskHandler{
		fileService: fileService,
	}
}

func (h *FileTaskHandler) HandlePdfProcessTask(ctx context.Context, t *asynq.Task) (err error) {
	var p tasks.PdfProcessPayload
	if err = json.Unmarshal(t.Payload(), &p); err != nil {
		logger.Logger.Error("Failed to parse task payload", zap.Error(err), zap.String("file_id", p.FileID))
		return err
	}

	logger.Logger.Info(
		"Starting extracting text content",
		zap.String("file_id", p.FileID),
		zap.String("user_id", p.UserID),
	)

	baseInput := services.UpdateFileInput{
		FileID: p.FileID,
		UserID: p.UserID,
	}

	defer func() {
		if err != nil && p.FileID != "" {
			input := baseInput
			input.Status = Ptr(domain.FileStatusError)
			_, updateErr := h.fileService.UpdateFile(context.Background(), input)
			if updateErr != nil {
				logger.Logger.Error(
					"Failed to update file status to error",
					zap.Error(updateErr),
					zap.String("file_id", p.FileID),
				)
			}
		}
	}()

	input := baseInput
	input.Status = Ptr(domain.FileStatusProcessing)
	file, err := h.fileService.UpdateFile(ctx, input)
	if err != nil {
		return err
	}

	text, err := h.fileService.ExtractTextFromFile(ctx, file)
	if err != nil {
		return err
	}

	input = baseInput
	input.TextContent = Ptr(text)
	file, err = h.fileService.UpdateFile(ctx, input)
	if err != nil {
		return err
	}

	input = baseInput
	input.Status = Ptr(domain.FileStatusReady)
	file, err = h.fileService.UpdateFile(ctx, input)
	if err != nil {
		return err
	}

	logger.Logger.Info(
		"Successfully extracted text content",
		zap.String("file_id", file.ID),
		zap.String("user_id", p.UserID),
	)

	return nil
}
