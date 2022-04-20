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

	core2 "github.com/open-component-model/ocm/pkg/contexts/credentials/core"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const KIND_CREDENTIALS = core2.KIND_CREDENTIALS
const KIND_CONSUMER = core2.KIND_CONSUMER
const KIND_REPOSITORY = core2.KIND_REPOSITORY

const CONTEXT_TYPE = core2.CONTEXT_TYPE

const AliasRepositoryType = core2.AliasRepositoryType

type Context = core2.Context
type RepositoryTypeScheme = core2.RepositoryTypeScheme
type Repository = core2.Repository
type Credentials = core2.Credentials
type CredentialsSource = core2.CredentialsSource
type CredentialsChain = core2.CredentialsChain
type CredentialsSpec = core2.CredentialsSpec
type RepositorySpec = core2.RepositorySpec

type ConsumerIdentity = core2.ConsumerIdentity
type IdentityMatcher = core2.IdentityMatcher

type GenericRepositorySpec = core2.GenericRepositorySpec
type GenericCredentialsSpec = core2.GenericCredentialsSpec
type DirectCredentials = core2.DirectCredentials

func DefaultContext() core2.Context {
	return core2.DefaultContext
}

func ForContext(ctx context.Context) Context {
	return core2.ForContext(ctx)
}

func NewCredentialsSpec(name string, repospec RepositorySpec) CredentialsSpec {
	return core2.NewCredentialsSpec(name, repospec)
}

func NewGenericCredentialsSpec(name string, repospec *GenericRepositorySpec) CredentialsSpec {
	return core2.NewGenericCredentialsSpec(name, repospec)
}

func NewGenericRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return core2.NewGenericRepositorySpec(data, unmarshaler)
}

func NewCredentials(props common.Properties) Credentials {
	return core2.NewCredentials(props)
}

func ToGenericCredentialsSpec(spec CredentialsSpec) (*GenericCredentialsSpec, error) {
	return core2.ToGenericCredentialsSpec(spec)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core2.ToGenericRepositorySpec(spec)
}

func ErrUnknownCredentials(name string) error {
	return core2.ErrUnknownCredentials(name)
}
