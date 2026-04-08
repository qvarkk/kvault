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
	"github.com/ledongthuc/pdf"
)

type FileTaskRepo interface {
	GetByID(context.Context, string) (*domain.File, error)
}

type FileTaskService struct {
	fileRepo FileTaskRepo
	aws      *aws.Aws
}

func NewFileTaskService(fileRepo FileTaskRepo, aws *aws.Aws) *FileTaskService {
	return &FileTaskService{
		fileRepo: fileRepo,
		aws:      aws,
	}
}

func (s *FileTaskService) ExtractTextFromS3(ctx context.Context, fileID string) (string, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return "", err
	}

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
