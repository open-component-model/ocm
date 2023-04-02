// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package actions

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"

	ocmcreds "github.com/open-component-model/ocm/pkg/contexts/credentials"
	oci_repository_prepare "github.com/open-component-model/ocm/pkg/contexts/datacontext/action/types/oci-repository-prepare"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
)

const defaultRegion = "us-west-1"

type Action struct{}

var _ ppi.Action = (*Action)(nil)

func (a Action) Name() string {
	return oci_repository_prepare.Type
}

func (a Action) Description() string {
	return "Create ECR repository if it does not yet exist."
}

func (a Action) Execute(p ppi.Plugin, spec ppi.ActionSpec, creds ocmcreds.DirectCredentials) (result ppi.ActionResult, err error) {
	prepare, ok := spec.(*oci_repository_prepare.ActionSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}

	ctx := context.Background()
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(defaultRegion),
	}

	var awsCred aws.CredentialsProvider = aws.AnonymousCredentials{}

	if creds != nil {
		accessKeyID := creds.GetProperty(ocmcreds.ATTR_AWS_ACCESS_KEY_ID)
		accessSecret := creds.GetProperty(ocmcreds.ATTR_AWS_SECRET_ACCESS_KEY)
		accessToken := creds.GetProperty(ocmcreds.ATTR_TOKEN)
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
		o.Region = defaultRegion
	})

	msg := fmt.Sprintf("repository %q already exists", prepare.Repository)
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
			return oci_repository_prepare.Result(fmt.Sprintf("repository %q created", prepare.Repository)), nil
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
