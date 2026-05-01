package services

import (
	"bytes"
	"context"
	"io"
	"os"
	"qvarkk/kvault/internal/aws"
	"qvarkk/kvault/internal/domain"
	"strings"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
	"github.com/ledongthuc/pdf"
)

type UpdateFileInput struct {
	FileID      string
	UserID      string
	Status      *domain.FileStatus
	TextContent *string
}

type FileTaskRepo interface {
	GetActiveByIDForUpdate(context.Context, *sqlx.Tx, string) (*domain.File, error)
	UpdateTx(context.Context, *sqlx.Tx, *domain.File) error
}

type FileTaskService struct {
	fileRepo   FileTaskRepo
	transactor Transactor
	aws        *aws.Aws
}

func NewFileTaskService(fileRepo FileTaskRepo, transactor Transactor, aws *aws.Aws) *FileTaskService {
	return &FileTaskService{
		fileRepo:   fileRepo,
		transactor: transactor,
		aws:        aws,
	}
}

func (s *FileTaskService) ExtractTextFromFile(ctx context.Context, file *domain.File) (string, error) {
	resp, err := s.aws.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: awsSdk.String(s.aws.BucketName),
		Key:    awsSdk.String(file.S3Key),
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", err
	}

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	fileInfo, err := tmpFile.Stat()
	if err != nil {
		return "", err
	}

	r, err := pdf.NewReader(tmpFile, fileInfo.Size())
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	_, err = buf.ReadFrom(b)
	if err != nil {
		return "", err
	}

	rawText := buf.String()
	words := strings.Fields(rawText)
	normalizedText := strings.Join(words, " ")
	return normalizedText, nil
}

func (s *FileTaskService) UpdateFile(
	ctx context.Context,
	input UpdateFileInput,
) (*domain.File, error) {
	var updated *domain.File

	err := s.transactor.WithTx(ctx, func(tx *sqlx.Tx) error {
		file, err := s.fileRepo.GetActiveByIDForUpdate(ctx, tx, input.FileID)
		if err != nil {
			return NewServiceError(ErrFileNotFound, "not found", err)
		}

		if file.UserID != input.UserID {
			return NewServiceError(ErrFileNotFound, "forbidden", err)
		}

		if input.Status != nil {
			file.Status = *input.Status
		}
		if input.TextContent != nil {
			file.TextContent = NewNullString(*input.TextContent)
		}

		if err := s.fileRepo.UpdateTx(ctx, tx, file); err != nil {
			return NewServiceError(ErrInternal, "update file internal error", err)
		}

		updated = file
		return nil
	})

	return updated, err
}
