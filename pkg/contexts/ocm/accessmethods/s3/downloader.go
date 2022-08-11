package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/open-component-model/ocm/pkg/common/accessio"
)

const defaultRegion = "us-west-1"

// Downloader defines a downloader for AWS S3 objects.
type Downloader interface {
	Download(region, bucket, key, version string, creds *AWSCreds) ([]byte, error)
}

// S3Downloader is a downloader capable of downloading S3 Objects.
type S3Downloader struct {
	cache accessio.BlobAccess
}

func NewS3Downloader(cache accessio.BlobAccess) *S3Downloader {
	return &S3Downloader{
		cache: cache,
	}
}

// AWSCreds groups AWS related credential values together.
type AWSCreds struct {
	AccessKeyID  string
	AccessSecret string
	SessionToken string
}

func (s *S3Downloader) Download(region, bucket, key, version string, creds *AWSCreds) ([]byte, error) {
	ctx := context.Background()
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	var awsCred aws.CredentialsProvider = aws.AnonymousCredentials{}
	if creds != nil {
		awsCred = awscreds.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     creds.AccessKeyID,
				SecretAccessKey: creds.AccessSecret,
			},
		}
	}
	opts = append(opts, config.WithCredentialsProvider(awsCred))
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration for AWS: %w", err)
	}

	if region == "" {
		var err error
		// deliberately use a different client so the real one will use the right region.
		// Region has to be provided to get the region of the specified bucket. We use the
		// global "default" of us-west-1 here. This will be updated to the right region
		// once we retrieve it or die trying.
		cfg.Region = defaultRegion
		region, err = manager.GetBucketRegion(context.Background(), s3.NewFromConfig(cfg), bucket, func(o *s3.Options) {
			o.Region = defaultRegion
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find bucket region: %w", err)
		}
		cfg.Region = region
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Credentials = awsCred
		o.Region = region
	})
	downloader := manager.NewDownloader(client)

	var blob []byte
	// instead of this, use the cache from Uwe.
	buf := manager.NewWriteAtBuffer(blob)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if version != "" {
		input.VersionId = aws.String(version)
	}
	if _, err := downloader.Download(context.Background(), buf, input); err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return buf.Bytes(), nil
}
