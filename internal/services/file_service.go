package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"qvarkk/kvault/internal/aws"
	"qvarkk/kvault/internal/domain"
	"qvarkk/kvault/internal/redis"
	"qvarkk/kvault/internal/tasks"
	"time"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

type FileRepo interface {
	CreateNew(context.Context, *domain.File) error
	List(context.Context, domain.ListFileParams) ([]domain.File, int, error)
	GetByID(context.Context, string) (*domain.File, error)
	GetActiveByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.File, error)
	GetDeletedByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.File, error)
	SoftDeleteByIDTx(context.Context, *sqlx.Tx, string) error
	RestoreByIDTx(context.Context, *sqlx.Tx, string) error
}

type FileService struct {
	fileRepo   FileRepo
	transactor Transactor
	redis      *redis.Redis
	aws        *aws.Aws
}

type CreateFileInput struct {
	UserID       string
	OriginalName string
	S3Key        string
	Size         int64
	MimeType     string
	Status       string
}

func NewFileService(fileRepo FileRepo, transactor Transactor, redis *redis.Redis, aws *aws.Aws) *FileService {
	return &FileService{
		fileRepo:   fileRepo,
		transactor: transactor,
		redis:      redis,
		aws:        aws,
	}
}

func (s *FileService) CreateNew(ctx context.Context, input CreateFileInput) (*domain.File, error) {
	file := &domain.File{
		UserID:       input.UserID,
		OriginalName: input.OriginalName,
		S3Key:        input.S3Key,
		Size:         input.Size,
		MimeType:     input.MimeType,
		Status:       domain.FileStatus(input.Status),
	}

	err := s.fileRepo.CreateNew(ctx, file)
	if err != nil {
		return nil, NewServiceError(ErrFileNotCreated, "database error", err)
	}

	return file, nil
}

func (s *FileService) List(ctx context.Context, params domain.ListFileParams) ([]domain.File, int, error) {
	files, count, err := s.fileRepo.List(ctx, params)
	if err != nil {
		return nil, 0, NewServiceError(ErrInternal, "list files internal error", err)
	}
	return files, count, err
}

func (s *FileService) GetFilePresignedUrl(ctx context.Context, fileID, userID string) (*domain.PresignedURL, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, NewServiceError(ErrFileNotFound, "not found", err)
	}

	if file.UserID != userID {
		return nil, NewServiceError(ErrFileNotFound, "forbidden", nil)
	}

	presignClient := s3.NewPresignClient(s.aws.S3Client)
	contentDispositionParam := fmt.Sprintf(
		"attachment; filename*=UTF-8''%s",
		url.PathEscape(file.OriginalName),
	)

	presignedResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     awsSdk.String(s.aws.BucketName),
		Key:                        awsSdk.String(file.S3Key),
		ResponseContentDisposition: awsSdk.String(contentDispositionParam),
	}, s3.WithPresignExpires(time.Second*time.Duration(s.aws.UrlExpirationTimeSeconds)))

	expiresAt := time.Now().UTC().Add(time.Second * time.Duration(s.aws.UrlExpirationTimeSeconds))

	return &domain.PresignedURL{
		URL:       presignedResult.URL,
		Filename:  file.OriginalName,
		MimeType:  file.MimeType,
		Size:      file.Size,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *FileService) DeleteByID(ctx context.Context, fileID, userID string) error {
	return s.authorizeAndMutateTx(
		ctx, fileID, userID,
		s.fileRepo.GetActiveByIDForUpdate,
		s.fileRepo.SoftDeleteByIDTx,
	)
}

func (s *FileService) RestoreByID(ctx context.Context, fileID, userID string) error {
	return s.authorizeAndMutateTx(
		ctx, fileID, userID,
		s.fileRepo.GetDeletedByIDForUpdate,
		s.fileRepo.RestoreByIDTx,
	)
}

func (s *FileService) authorizeAndMutateTx(
	ctx context.Context, fileID, userID string,
	getFn func(context.Context, *sqlx.Tx, string) (*domain.File, error),
	mutateFn func(context.Context, *sqlx.Tx, string) error,
) error {
	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		file, err := getFn(ctx, tx, fileID)
		if err != nil {
			return NewServiceError(ErrFileNotFound, "not found", err)
		}

		if file.UserID != userID {
			return NewServiceError(ErrFileNotFound, "forbidden", nil)
		}

		err = mutateFn(ctx, tx, fileID)
		if err != nil {
			return NewServiceError(ErrInternal, "mutate file internal error", err)
		}

		return nil
	})

	return err
}

func (s *FileService) ValidatePdfFile(ctx context.Context, fileHeader *multipart.FileHeader) error {
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

func (s *FileService) UploadPdfFileToS3(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := uuid.New().String() + ".pdf"
	_, err = s.aws.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: awsSdk.String(s.aws.BucketName),
		Key:    awsSdk.String(s.aws.GetKey(filename)),
		Body:   file,
	})

	return s.aws.GetKey(filename), err
}

func (s *FileService) EnqueuePdfProcessTask(ctx context.Context, payload tasks.PdfProcessPayload) (*asynq.TaskInfo, error) {
	task, err := tasks.NewPdfProcessTask(payload)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to create PDF processing task", err)
	}

	info, err := s.redis.AsynqClient.EnqueueContext(ctx, task)
	if err != nil {
		return nil, NewServiceError(ErrInternal, "failed to enqueue PDF processing task", err)
	}

	return info, nil
}
