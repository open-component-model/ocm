package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Downloader defines a downloader for AWS S3 objects.
type Downloader interface {
	Download(region, bucket, key, version, accessKeyID, accessSecret string) ([]byte, error)
}

type S3Downloader struct {
}

func (s *S3Downloader) Download(region, bucket, key, version, accessKeyID, accessSecret string) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region), config.WithCredentialsProvider(awscreds.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: accessSecret,
			//SessionToken:    "", // TODO: come back to this
		},
	}))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	downloader := manager.NewDownloader(client)
	ctx := context.Background()

	var blob []byte
	buf := manager.NewWriteAtBuffer(blob)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if version != "" {
		input.VersionId = aws.String(version)
	}
	if _, err := downloader.Download(ctx, buf, input); err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return buf.Bytes(), nil
}
