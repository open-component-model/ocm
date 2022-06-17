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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const CONTEXT_TYPE = core.CONTEXT_TYPE

const CommonTransportFormat = core.CommonTransportFormat

type Context = core.Context
type ComponentVersionResolver = core.ComponentVersionResolver
type Repository = core.Repository
type RepositorySpecHandlers = core.RepositorySpecHandlers
type RepositorySpecHandler = core.RepositorySpecHandler
type UniformRepositorySpec = core.UniformRepositorySpec
type ComponentLister = core.ComponentLister
type ComponentAccess = core.ComponentAccess
type ComponentVersionAccess = core.ComponentVersionAccess
type AccessSpec = core.AccessSpec
type GenericAccessSpec = core.GenericAccessSpec
type HintProvider = core.HintProvider
type AccessMethod = core.AccessMethod
type AccessMethodSupport = core.AccessMethodSupport
type AccessType = core.AccessType
type DataAccess = core.DataAccess
type BlobAccess = core.BlobAccess
type SourceAccess = core.SourceAccess
type SourceMeta = core.SourceMeta
type ResourceAccess = core.ResourceAccess
type ResourceMeta = core.ResourceMeta
type RepositorySpec = core.RepositorySpec
type IntermediateRepositorySpecAspect = core.IntermediateRepositorySpecAspect
type GenericRepositorySpec = core.GenericRepositorySpec
type RepositoryType = core.RepositoryType
type ComponentReference = core.ComponentReference

type BlobHandler = core.BlobHandler
type BlobHandlerOption = core.BlobHandlerOption
type StorageContext = core.StorageContext
type ImplementationRepositoryType = core.ImplementationRepositoryType

type DigesterType = core.DigesterType
type BlobDigester = core.BlobDigester
type BlobDigesterRegistry = core.BlobDigesterRegistry
type DigestDescriptor = core.DigestDescriptor

func New() Context {
	return core.Builder{}.New()
}

func NewDigestDescriptor(digest string, typ DigesterType) *DigestDescriptor {
	return core.NewDigestDescriptor(digest, typ.HashAlgorithm, typ.NormalizationAlgorithm)
}

func DefaultBlobDigesterRegistry() BlobDigesterRegistry {
	return core.DefaultBlobDigesterRegistry
}

func DefaultContext() core.Context {
	return core.DefaultContext
}

func WithPrio(p int) BlobHandlerOption {
	return core.WithPrio(p)
}

func ForRepo(ctxtype, repostype string) BlobHandlerOption {
	return core.ForRepo(ctxtype, repostype)
}

func ForMimeType(mimetype string) BlobHandlerOption {
	return core.ForMimeType(mimetype)
}

func RegisterRepositorySpecHandler(handler RepositorySpecHandler, types ...string) {
	core.RegisterRepositorySpecHandler(handler, types...)
}

func RegisterBlobHandler(handler BlobHandler, opts ...BlobHandlerOption) {
	core.RegisterBlobHandler(handler, opts...)
}

func RegisterRepositoryType(name string, atype RepositoryType) {
	core.DefaultRepositoryTypeScheme.Register(name, atype)
}

func RegisterAccessType(atype AccessType) {
	core.DefaultAccessTypeScheme.Register(atype.GetKind(), atype)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core.ToGenericRepositorySpec(spec)
}

type AccessSpecRef = core.AccessSpecRef

func NewAccessSpecRef(spec AccessSpec) *AccessSpecRef {
	return core.NewAccessSpecRef(spec)
}

func NewRawAccessSpecRef(data []byte, unmarshaler runtime.Unmarshaler) (*AccessSpecRef, error) {
	return core.NewRawAccessSpecRef(data, unmarshaler)
}

const KIND_COMPONENTVERSION = core.KIND_COMPONENTVERSION
const KIND_RESOURCE = core.KIND_RESOURCE
const KIND_SOURCE = core.KIND_SOURCE
const KIND_REFERENCE = core.KIND_REFERENCE

func ErrComponentVersionNotFound(name, version string) error {
	return core.ErrComponentVersionNotFound(name, version)
}

func ErrComponentVersionNotFoundWrap(err error, name, version string) error {
	return core.ErrComponentVersionNotFoundWrap(err, name, version)
}
