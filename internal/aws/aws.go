package aws

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"qvarkk/kvault/config"
	"time"

	awslib "github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	uploadsPrefix = "uploads/"
)

type AwsStorage struct {
	s3Client          *s3.Client
	presignClient     *s3.PresignClient
	bucketName        string
	prefix            string
	urlExpiration     time.Duration
	viewUrlExpiration time.Duration
}

func NewAwsStorage(config config.AwsConfig) (*AwsStorage, error) {
	awsConfig, err := awscfg.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	presignBase := client
	if config.PublicEndpointUrl != "" {
		presignBase = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			o.UsePathStyle = true
			o.BaseEndpoint = awslib.String(config.PublicEndpointUrl)
		})
	}

	storage := &AwsStorage{
		s3Client:          client,
		presignClient:     s3.NewPresignClient(presignBase),
		bucketName:        config.S3Bucket,
		prefix:            uploadsPrefix,
		urlExpiration:     config.UrlExpiration,
		viewUrlExpiration: config.ViewUrlExpiration,
	}

	if err := storage.setupCors(context.TODO()); err != nil {
		return nil, fmt.Errorf("setup bucket CORS: %w", err)
	}

	return storage, nil
}

func (s *AwsStorage) setupCors(ctx context.Context) error {
	_, err := s.s3Client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: awslib.String(s.bucketName),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: []types.CORSRule{
				{
					AllowedOrigins: []string{"*"},
					AllowedMethods: []string{"GET", "HEAD"},
					AllowedHeaders: []string{"*"},
				},
			},
		},
	})
	return err
}

func (s *AwsStorage) Upload(
	ctx context.Context,
	key string,
	body io.Reader,
) error {
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: awslib.String(s.bucketName),
		Key:    awslib.String(s.fullKey(key)),
		Body:   body,
	})
	return err
}

func (s *AwsStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: awslib.String(s.bucketName),
		Key:    awslib.String(s.fullKey(key)),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s *AwsStorage) Delete(ctx context.Context, key string) error {
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: awslib.String(s.bucketName),
		Key:    awslib.String(s.fullKey(key)),
	})
	return err
}

func (s *AwsStorage) GeneratePresignUrl(
	ctx context.Context,
	key, filename string,
) (string, time.Time, error) {
	contentDispositionParam := fmt.Sprintf(
		"attachment; filename*=UTF-8''%s",
		url.PathEscape(filename),
	)

	presignedResult, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     awslib.String(s.bucketName),
		Key:                        awslib.String(s.fullKey(key)),
		ResponseContentDisposition: awslib.String(contentDispositionParam),
	}, s3.WithPresignExpires(s.urlExpiration))
	if err != nil {
		return "", time.Time{}, err
	}

	return presignedResult.URL, time.Now().UTC().Add(s.urlExpiration), nil
}

func (s *AwsStorage) GeneratePresignViewUrl(
	ctx context.Context,
	key string,
) (string, time.Time, error) {
	presignedResult, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     awslib.String(s.bucketName),
		Key:                        awslib.String(s.fullKey(key)),
		ResponseContentDisposition: awslib.String("inline"),
	}, s3.WithPresignExpires(s.viewUrlExpiration))
	if err != nil {
		return "", time.Time{}, err
	}

	return presignedResult.URL, time.Now().UTC().Add(s.viewUrlExpiration), nil
}

func (s *AwsStorage) fullKey(key string) string {
	return s.prefix + key
}
