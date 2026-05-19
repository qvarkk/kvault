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
	publicEndpointUrl string
}

func NewAwsStorage(config config.AwsConfig) (*AwsStorage, error) {
	awsConfig, err := awscfg.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &AwsStorage{
		s3Client:          client,
		presignClient:     s3.NewPresignClient(client),
		bucketName:        config.S3Bucket,
		prefix:            uploadsPrefix,
		urlExpiration:     config.UrlExpiration,
		publicEndpointUrl: config.PublicEndpointUrl,
	}, nil
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

	resultURL := presignedResult.URL
	if s.publicEndpointUrl != "" {
		parsed, err := url.Parse(presignedResult.URL)
		if err == nil {
			pub, err := url.Parse(s.publicEndpointUrl)
			if err == nil {
				parsed.Scheme = pub.Scheme
				parsed.Host = pub.Host
				resultURL = parsed.String()
			}
		}
	}

	return resultURL, time.Now().UTC().Add(s.urlExpiration), nil
}

func (s *AwsStorage) fullKey(key string) string {
	return s.prefix + key
}
