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
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
)

const KIND_CREDENTIALS = core.KIND_CREDENTIALS
const KIND_REPOSITORY = core.KIND_REPOSITORY

const CONTEXT_TYPE = core.CONTEXT_TYPE

type Context = core.Context
type Repository = core.Repository
type RepositoryType = core.RepositoryType
type Credentials = core.Credentials
type CredentialsSource = core.CredentialsSource
type CredentialsChain = core.CredentialsChain
type CredentialsSpec = core.CredentialsSpec
type RepositorySpec = core.RepositorySpec
type GenericRepositorySpec = core.GenericRepositorySpec
type GenericCredentialsSpec = core.GenericCredentialsSpec

type ConsumerIdentity = core.ConsumerIdentity
type IdentityMatcher = core.IdentityMatcher

var DefaultContext = core.DefaultContext

func NewGenericCredentialsSpec(name string, repospec *GenericRepositorySpec) *GenericCredentialsSpec {
	return core.NewGenericCredentialsSpec(name, repospec)
}

func NewCredentialsSpec(name string, repospec RepositorySpec) CredentialsSpec {
	return core.NewCredentialsSpec(name, repospec)
}

func ToGenericCredentialsSpec(spec CredentialsSpec) (*GenericCredentialsSpec, error) {
	return core.ToGenericCredentialsSpec(spec)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core.ToGenericRepositorySpec(spec)
}

func RegisterRepositoryType(name string, atype RepositoryType) {
	core.DefaultRepositoryTypeScheme.Register(name, atype)
}

func RegisterIdentityMatcher(typ string, matcher IdentityMatcher, desc string) {
	core.StandardIdentityMatchers.Register(typ, matcher, desc)
}

func NewCredentials(props common.Properties) Credentials {
	return core.NewCredentials(props)
}

func ErrUnknownCredentials(name string) error {
	return core.ErrUnknownCredentials(name)
}

func ErrUnknownRepository(kind, name string) error {
	return core.ErrUnknownRepository(kind, name)
}

var CompleteMatch = core.CompleteMatch
var NoMatch = core.NoMatch
var PartialMatch = core.PartialMatch
