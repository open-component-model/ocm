package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/mandelsoft/goutils/errors"
	ocmcreds "ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci/extensions/actions/oci-repository-prepare"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/s3/identity"
	"ocm.software/ocm/api/ocm/plugin/ppi"
)

type Action struct{}

var _ ppi.Action = (*Action)(nil)

func (a Action) Name() string {
	return oci_repository_prepare.Type
}

func (a Action) Description() string {
	return "Create ECR repository if it does not yet exist."
}

func (a Action) ConsumerType() string {
	return "AWS"
}

func (a Action) Execute(p ppi.Plugin, spec ppi.ActionSpec, creds ocmcreds.DirectCredentials) (result ppi.ActionResult, err error) {
	prepare, ok := spec.(*oci_repository_prepare.ActionSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}

	path := strings.Split(prepare.Hostname, ".")
	if len(path) < 5 {
		return nil, fmt.Errorf("unknown ecr host %q", prepare.Hostname)
	}
	if path[len(path)-4] != "ecr" {
		return nil, fmt.Errorf("unknown ecr host %q", prepare.Hostname)
	}
	region := path[len(path)-3]
	ctx := context.Background()
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	var awsCred aws.CredentialsProvider = aws.AnonymousCredentials{}

	if creds != nil {
		accessKeyID := creds.GetProperty(identity.ATTR_AWS_ACCESS_KEY_ID)
		accessSecret := creds.GetProperty(identity.ATTR_AWS_SECRET_ACCESS_KEY)
		accessToken := creds.GetProperty(identity.ATTR_TOKEN)
		awsCred = credentials.NewStaticCredentialsProvider(accessKeyID, accessSecret, accessToken)
	}

	opts = append(opts, config.WithCredentialsProvider(awsCred))
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration for AWS: %w", err)
	}

	client := ecr.NewFromConfig(cfg, func(o *ecr.Options) {
		// Pass in creds because of https://github.com/aws/aws-sdk-go-v2/issues/1797
		o.Credentials = awsCred
		o.Region = region
	})

	msg := fmt.Sprintf("repository %q already exists in region %s", prepare.Repository, region)
	in := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{prepare.Repository},
	}
	_, err = client.DescribeRepositories(ctx, in)
	if err != nil {
		var rnf *types.RepositoryNotFoundException
		if errors.As(err, &rnf) {
			in := &ecr.CreateRepositoryInput{
				RepositoryName: aws.String(prepare.Repository),
				Tags: []types.Tag{
					{
						Key:   aws.String("ocm"),
						Value: aws.String("ecrplugin"),
					},
				},
			}
			_, err := client.CreateRepository(ctx, in)
			if err != nil {
				var re *types.RepositoryAlreadyExistsException
				if errors.As(err, &re) {
					return oci_repository_prepare.Result(msg), nil
				}
				return nil, err
			}
			return oci_repository_prepare.Result(fmt.Sprintf("repository %q created in region %s", prepare.Repository, region)), nil
		}
		return nil, err
	}
	return oci_repository_prepare.Result(msg), nil
}

func (a *Action) DefaultSelectors() []string {
	return []string{".*\\.dkr\\.ecr\\..*\\.amazonaws\\.com"}
}

func New() ppi.Action {
	return &Action{}
}
