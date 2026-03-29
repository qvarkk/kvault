package services

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/redis"
	"qvarkk/kvault/internal/tasks"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ItemRepo interface {
	CreateNew(context.Context, *domain.Item) error
	CreateFileMeta(context.Context, *domain.FileMeta) error
}

type ItemService struct {
	itemRepo ItemRepo
	redis    *redis.Redis
}

type CreateItemInput struct {
	UserID     string
	Type       string
	Title      string
	Content    string
	FileMetaID string
}

type CreateFileMetaInput struct {
	Path     string
	Size     int64
	MimeType string
	Status   string
}

func NewItemService(itemRepo ItemRepo, redis *redis.Redis) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
		redis:    redis,
	}
}

func (i *ItemService) CreateNew(ctx context.Context, input CreateItemInput) (*domain.Item, error) {
	item := &domain.Item{
		UserID:     input.UserID,
		Type:       domain.ItemType(input.Type),
		Title:      input.Title,
		Content:    NewNullString(input.Content),
		FileMetaID: NewNullString(input.FileMetaID),
	}

	err := i.itemRepo.CreateNew(ctx, item)
	if err != nil {
		return nil, NewServiceError(ErrItemNotCreated, "database error", err)
	}

	return item, nil
}

func (i *ItemService) CreateFileMeta(ctx context.Context, input CreateFileMetaInput) (*domain.FileMeta, error) {
	fileMeta := &domain.FileMeta{
		Path:     input.Path,
		Size:     input.Size,
		MimeType: input.MimeType,
		Status:   domain.FileStatus(input.Status),
	}

	err := i.itemRepo.CreateFileMeta(ctx, fileMeta)
	if err != nil {
		return nil, NewServiceError(ErrFileMetaNotCreated, "database error", err)
	}

	return fileMeta, nil
}

func (i *ItemService) ValidatePdfFile(ctx context.Context, fileHeader *multipart.FileHeader) error {
	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".pdf" {
		return NewServiceError(ErrPdfFileFormat, "invalid file extension", nil)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return NewServiceError(ErrInternal, "failed to open uploaded file", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return NewServiceError(ErrInternal, "failed to read uploaded file", err)
	}

	contentType := http.DetectContentType(buffer[:n])
	if contentType != "application/pdf" {
		return NewServiceError(ErrPdfFileFormat, "File should be of a PDF content type.", nil)
	}

	return nil
}

func (i *ItemService) GeneratePdfFileDestination() string {
	filename := uuid.New().String() + ".pdf"
	return filepath.Join("./tmp/uploads", filename)
}

func (i *ItemService) EnqueueFileUploadTask(ctx context.Context, payload tasks.FileUploadPayload) (*asynq.TaskInfo, error) {
	task, err := tasks.NewFileUploadTask(payload)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to create task", err)
	}

	info, err := i.redis.AsynqClient.EnqueueContext(ctx, task)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to enqueue task", err)
	}

	return info, nil
}
