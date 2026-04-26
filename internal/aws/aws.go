package aws

import (
	"context"
	"path/filepath"
	"qvarkk/kvault/config"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	uploadsPrefix = "uploads/"
)

type Aws struct {
	S3Client                 *s3.Client
	BucketName               string
	Prefix                   string
	UrlExpirationTimeSeconds int
}

func (a *Aws) GetKey(filename string) string {
	return filepath.Join(a.Prefix, filename)
}

func NewAws(config config.AwsConfig) (*Aws, error) {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &Aws{
		S3Client:                 client,
		BucketName:               config.S3Bucket,
		Prefix:                   uploadsPrefix,
		UrlExpirationTimeSeconds: config.UrlExpirationTimeSeconds,
	}, nil
}
