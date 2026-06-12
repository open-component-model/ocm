package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const defaultRegion = "us-west-1"

// Downloader is a downloader capable of downloading S3 Objects.
type Downloader struct {
	region, bucket, key, version string
	creds                        *AWSCreds
}

func NewDownloader(region, bucket, key, version string, creds *AWSCreds) *Downloader {
	return &Downloader{
		region:  region,
		bucket:  bucket,
		key:     key,
		version: version,
		creds:   creds,
	}
}

// AWSCreds groups AWS related credential values together.
type AWSCreds struct {
	AccessKeyID  string
	AccessSecret string
	SessionToken string
}

func (s *Downloader) Download(w io.WriterAt) error {
	ctx := context.Background()
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(s.region),
	}
	var awsCred aws.CredentialsProvider = aws.AnonymousCredentials{}
	if s.creds != nil {
		awsCred = awscreds.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     s.creds.AccessKeyID,
				SecretAccessKey: s.creds.AccessSecret,
			},
		}
	}
	opts = append(opts, config.WithCredentialsProvider(awsCred))
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load configuration for AWS: %w", err)
	}

	if s.region == "" {
		// deliberately use a different client so the real one will use the right region.
		// Region has to be provided to get the region of the specified bucket. We use the
		// global "default" of us-west-1 here. This will be updated to the right region
		// once we retrieve it or die trying.
		// With the new API introduced, transfermanager no longer has GetBucketRegion.
		// Thus, we just implement GetBucketRegion here instead as it was in the old manager SDK.
		// construct the default client with the default region
		tmpClient := s3.NewFromConfig(cfg, func(o *s3.Options) {
			// Pass in creds because of https://github.com/aws/aws-sdk-go-v2/issues/1797
			o.Credentials = awsCred
			o.Region = defaultRegion
		})
		resp, err := tmpClient.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(s.bucket),
		})
		if err == nil {
			s.region = aws.ToString(resp.BucketRegion)
		} else {
			// S3 returns 301 when the bucket is in a different region than the hint region.
			// The SDK does not follow 301 redirects, but the correct region is in X-Amz-Bucket-Region.
			var respErr *smithyhttp.ResponseError
			if !errors.As(err, &respErr) || respErr.HTTPStatusCode() != http.StatusMovedPermanently {
				return fmt.Errorf("failed to find bucket region: %w", err)
			}
			region := respErr.Response.Header.Get("X-Amz-Bucket-Region")
			if region == "" {
				return fmt.Errorf("failed to find bucket region: %w", err)
			}
			s.region = region
		}
		cfg.Region = s.region
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// Pass in creds because of https://github.com/aws/aws-sdk-go-v2/issues/1797
		o.Credentials = awsCred
		o.Region = s.region
	})

	downloader := transfermanager.New(client)
	input := &transfermanager.DownloadObjectInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(s.key),
		WriterAt: w,
	}
	if s.version != "" {
		input.VersionID = aws.String(s.version)
	}

	if _, err := downloader.DownloadObject(ctx, input); err != nil {
		return fmt.Errorf("failed to download object: %w", err)
	}

	return nil
}
