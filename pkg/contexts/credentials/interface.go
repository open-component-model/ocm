// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
