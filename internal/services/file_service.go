package services

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"qvarkk/kvault/internal/aws"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/redis"
	"qvarkk/kvault/internal/tasks"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type FileRepo interface {
	CreateNew(context.Context, *domain.File) error
}

type FileService struct {
	fileRepo FileRepo
	redis    *redis.Redis
	aws      *aws.Aws
}

type CreateFileInput struct {
	UserID       string
	OriginalName string
	S3Key        string
	Size         int64
	MimeType     string
	Status       string
}

func NewFileService(fileRepo FileRepo, redis *redis.Redis, aws *aws.Aws) *FileService {
	return &FileService{
		fileRepo: fileRepo,
		redis:    redis,
		aws:      aws,
	}
}

func (f *FileService) CreateNew(ctx context.Context, input CreateFileInput) (*domain.File, error) {
	file := &domain.File{
		UserID:       input.UserID,
		OriginalName: input.OriginalName,
		S3Key:        input.S3Key,
		Size:         input.Size,
		MimeType:     input.MimeType,
		Status:       domain.FileStatus(input.Status),
	}

	err := f.fileRepo.CreateNew(ctx, file)
	if err != nil {
		return nil, NewServiceError(ErrFileNotCreated, "database error", err)
	}

	return file, nil
}

func (f *FileService) ValidatePdfFile(ctx context.Context, fileHeader *multipart.FileHeader) error {
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
		return NewServiceError(ErrPdfFileFormat, "invalid file content type", nil)
	}

	return nil
}

func (f *FileService) UploadPdfFileToS3(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := uuid.New().String() + ".pdf"
	_, err = f.aws.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: awsSdk.String(f.aws.BucketName),
		Key:    awsSdk.String(f.aws.GetKey(filename)),
		Body:   file,
	})

	return f.aws.GetKey(filename), err
}

func (f *FileService) EnqueuePdfProcessTask(ctx context.Context, payload tasks.PdfProcessPayload) (*asynq.TaskInfo, error) {
	task, err := tasks.NewPdfProcessTask(payload)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to create PDF processing task", err)
	}

	info, err := f.redis.AsynqClient.EnqueueContext(ctx, task)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to enqueue PDF processing task", err)
	}

	return info, nil
}
