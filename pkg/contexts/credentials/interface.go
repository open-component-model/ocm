// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	KIND_CREDENTIALS = core.KIND_CREDENTIALS
	KIND_CONSUMER    = core.KIND_CONSUMER
	KIND_REPOSITORY  = core.KIND_REPOSITORY
)

const CONTEXT_TYPE = core.CONTEXT_TYPE

const AliasRepositoryType = core.AliasRepositoryType

type (
	Context              = core.Context
	RepositoryTypeScheme = core.RepositoryTypeScheme
	Repository           = core.Repository
	Credentials          = core.Credentials
	CredentialsSource    = core.CredentialsSource
	CredentialsChain     = core.CredentialsChain
	CredentialsSpec      = core.CredentialsSpec
	RepositorySpec       = core.RepositorySpec
)

type (
	ConsumerIdentity        = core.ConsumerIdentity
	IdentityMatcher         = core.IdentityMatcher
	IdentityMatcherInfo     = core.IdentityMatcherInfo
	IdentityMatcherRegistry = core.IdentityMatcherRegistry
)

type (
	GenericRepositorySpec  = core.GenericRepositorySpec
	GenericCredentialsSpec = core.GenericCredentialsSpec
	DirectCredentials      = core.DirectCredentials
)

func DefaultContext() core.Context {
	return core.DefaultContext
}

func ForContext(ctx context.Context) Context {
	return core.ForContext(ctx)
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	return core.DefinedForContext(ctx)
}

func NewCredentialsSpec(name string, repospec RepositorySpec) CredentialsSpec {
	return core.NewCredentialsSpec(name, repospec)
}

func NewGenericCredentialsSpec(name string, repospec *GenericRepositorySpec) CredentialsSpec {
	return core.NewGenericCredentialsSpec(name, repospec)
}

func NewGenericRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return core.NewGenericRepositorySpec(data, unmarshaler)
}

func NewCredentials(props common.Properties) Credentials {
	return core.NewCredentials(props)
}

func ToGenericCredentialsSpec(spec CredentialsSpec) (*GenericCredentialsSpec, error) {
	return core.ToGenericCredentialsSpec(spec)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core.ToGenericRepositorySpec(spec)
}

func ErrUnknownCredentials(name string) error {
	return core.ErrUnknownCredentials(name)
}

var (
	CompleteMatch = core.CompleteMatch
	NoMatch       = core.NoMatch
	PartialMatch  = core.PartialMatch
)
