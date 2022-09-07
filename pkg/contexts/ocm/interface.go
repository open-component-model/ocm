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

package ocm

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	KIND_COMPONENTVERSION   = core.KIND_COMPONENTVERSION
	KIND_COMPONENTREFERENCE = "component reference"
	KIND_RESOURCE           = core.KIND_RESOURCE
	KIND_SOURCE             = core.KIND_SOURCE
	KIND_REFERENCE          = core.KIND_REFERENCE
)

const CONTEXT_TYPE = core.CONTEXT_TYPE

const CommonTransportFormat = core.CommonTransportFormat

type (
	Context                          = core.Context
	ComponentVersionResolver         = core.ComponentVersionResolver
	Repository                       = core.Repository
	RepositorySpecHandlers           = core.RepositorySpecHandlers
	RepositorySpecHandler            = core.RepositorySpecHandler
	UniformRepositorySpec            = core.UniformRepositorySpec
	ComponentLister                  = core.ComponentLister
	ComponentAccess                  = core.ComponentAccess
	ComponentVersionAccess           = core.ComponentVersionAccess
	AccessSpec                       = core.AccessSpec
	HintProvider                     = core.HintProvider
	AccessMethod                     = core.AccessMethod
	AccessType                       = core.AccessType
	DataAccess                       = core.DataAccess
	BlobAccess                       = core.BlobAccess
	SourceAccess                     = core.SourceAccess
	SourceMeta                       = core.SourceMeta
	ResourceAccess                   = core.ResourceAccess
	ResourceMeta                     = core.ResourceMeta
	RepositorySpec                   = core.RepositorySpec
	IntermediateRepositorySpecAspect = core.IntermediateRepositorySpecAspect
	RepositoryType                   = core.RepositoryType
	RepositoryTypeScheme             = core.RepositoryTypeScheme
	AccessTypeScheme                 = core.AccessTypeScheme
	ComponentReference               = core.ComponentReference
)

type (
	DigesterType         = core.DigesterType
	BlobDigester         = core.BlobDigester
	BlobDigesterRegistry = core.BlobDigesterRegistry
	DigestDescriptor     = core.DigestDescriptor
)

type (
	BlobHandlerRegistry = core.BlobHandlerRegistry
	BlobHandler         = core.BlobHandler
)

func NewDigestDescriptor(digest, hashAlgo, normAlgo string) *DigestDescriptor {
	return core.NewDigestDescriptor(digest, hashAlgo, normAlgo)
}

// DefaultContext is the default context initialized by init functions.
func DefaultContext() core.Context {
	return core.DefaultContext
}

func DefaultBlobHandlers() core.BlobHandlerRegistry {
	return core.DefaultBlobHandlerRegistry
}

// ForContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
func ForContext(ctx context.Context) Context {
	return core.ForContext(ctx)
}

func NewGenericAccessSpec(spec string) (AccessSpec, error) {
	return core.NewGenericAccessSpec(spec)
}

type AccessSpecRef = core.AccessSpecRef

func NewAccessSpecRef(spec cpi.AccessSpec) *AccessSpecRef {
	return core.NewAccessSpecRef(spec)
}

func NewRawAccessSpecRef(data []byte, unmarshaler runtime.Unmarshaler) (*AccessSpecRef, error) {
	return core.NewRawAccessSpecRef(data, unmarshaler)
}
