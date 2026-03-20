package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	bucketName := "kvault-bucket"
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("AWS_ENDPOINT_URL"))
		o.UsePathStyle = true
	})

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("go.mod"),
	})
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(output.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
