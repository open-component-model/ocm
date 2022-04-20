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

package cpi

// This is the Context Provider Interface for credential providers

import (
	"github.com/open-component-model/ocm/pkg/common"
	core2 "github.com/open-component-model/ocm/pkg/contexts/credentials/core"
)

const KIND_CREDENTIALS = core2.KIND_CREDENTIALS
const KIND_REPOSITORY = core2.KIND_REPOSITORY

const CONTEXT_TYPE = core2.CONTEXT_TYPE

type Context = core2.Context
type Repository = core2.Repository
type RepositoryType = core2.RepositoryType
type Credentials = core2.Credentials
type CredentialsSource = core2.CredentialsSource
type CredentialsChain = core2.CredentialsChain
type CredentialsSpec = core2.CredentialsSpec
type RepositorySpec = core2.RepositorySpec
type GenericRepositorySpec = core2.GenericRepositorySpec
type GenericCredentialsSpec = core2.GenericCredentialsSpec

type ConsumerIdentity = core2.ConsumerIdentity
type IdentityMatcher = core2.IdentityMatcher

var DefaultContext = core2.DefaultContext

func NewGenericCredentialsSpec(name string, repospec *GenericRepositorySpec) *GenericCredentialsSpec {
	return core2.NewGenericCredentialsSpec(name, repospec)
}

func NewCredentialsSpec(name string, repospec RepositorySpec) CredentialsSpec {
	return core2.NewCredentialsSpec(name, repospec)
}

func ToGenericCredentialsSpec(spec CredentialsSpec) (*GenericCredentialsSpec, error) {
	return core2.ToGenericCredentialsSpec(spec)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core2.ToGenericRepositorySpec(spec)
}

func RegisterRepositoryType(name string, atype RepositoryType) {
	core2.DefaultRepositoryTypeScheme.Register(name, atype)
}

func NewCredentials(props common.Properties) Credentials {
	return core2.NewCredentials(props)
}

func ErrUnknownCredentials(name string) error {
	return core2.ErrUnknownCredentials(name)
}

func ErrUnknownRepository(kind, name string) error {
	return core2.ErrUnknownRepository(kind, name)
}
