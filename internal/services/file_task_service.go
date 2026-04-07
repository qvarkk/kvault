package services

import (
	"bytes"
	"context"
	"qvarkk/kvault/internal/aws"
	"qvarkk/kvault/internal/domain"

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

func (s *FileTaskService) GetPdfFileFromS3(ctx context.Context, fileID string) (*bytes.Buffer, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	resp, err := s.aws.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: awsSdk.String(s.aws.BucketName),
		Key:    awsSdk.String(file.S3Key),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf, nil
}

func (s *FileTaskService) ConvertPDFToPlainText(ctx context.Context, reader *bytes.Reader) (string, error) {
	r, err := pdf.NewReader(reader, reader.Size())
	if err != nil {
		return "", nil
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	_, err = buf.ReadFrom(b)
	return buf.String(), err
}
