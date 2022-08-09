package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Downloader defines a downloader for AWS S3 objects.
type Downloader interface {
	Download(region, bucket, key, version string, creds *credentials.Credentials) ([]byte, error)
}

type S3Downloader struct {
}

func (s *S3Downloader) Download(region, bucket, key, version string, creds *credentials.Credentials) ([]byte, error) {

	cfg := &aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	}
	ctx := context.Background()
	sess, _ := session.NewSession(cfg)

	if region == "" {
		cfg.Region = aws.String("us-east-1")
		reg, err := s3manager.GetBucketRegion(ctx, sess, bucket, *cfg.Region)
		if err != nil {
			return nil, err
		}
		cfg.Region = aws.String(reg)
	}

	sess, _ = session.NewSession(cfg)
	downloader := s3manager.NewDownloader(sess)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if version != "" {
		input.VersionId = aws.String(version)
	}

	var blob []byte
	buf := aws.NewWriteAtBuffer(blob)

	if _, err := downloader.DownloadWithContext(ctx, buf, input); err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return buf.Bytes(), nil
}
