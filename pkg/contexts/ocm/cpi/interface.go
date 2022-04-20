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
	core2 "github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/opencontainers/go-digest"
)

const CONTEXT_TYPE = core2.CONTEXT_TYPE

const CommonTransportFormat = core2.CommonTransportFormat

type Context = core2.Context
type Repository = core2.Repository
type RepositorySpecHandlers = core2.RepositorySpecHandlers
type RepositorySpecHandler = core2.RepositorySpecHandler
type UniformRepositorySpec = core2.UniformRepositorySpec
type ComponentLister = core2.ComponentLister
type ComponentAccess = core2.ComponentAccess
type ComponentVersionAccess = core2.ComponentVersionAccess
type AccessSpec = core2.AccessSpec
type AccessMethod = core2.AccessMethod
type AccessMethodSupport = core2.AccessMethodSupport
type AccessType = core2.AccessType
type DataAccess = core2.DataAccess
type BlobAccess = core2.BlobAccess
type SourceAccess = core2.SourceAccess
type SourceMeta = core2.SourceMeta
type ResourceAccess = core2.ResourceAccess
type ResourceMeta = core2.ResourceMeta
type RepositorySpec = core2.RepositorySpec
type GenericRepositorySpec = core2.GenericRepositorySpec
type RepositoryType = core2.RepositoryType
type ComponentReference = core2.ComponentReference

type BlobHandler = core2.BlobHandler
type BlobHandlerKey = core2.BlobHandlerKey
type StorageContext = core2.StorageContext

type DigesterType = core2.DigesterType
type BlobDigester = core2.BlobDigester
type BlobDigesterRegistry = core2.BlobDigesterRegistry
type DigestDescriptor = core2.DigestDescriptor

func NewDigestDescriptor(digest digest.Digest, typ DigesterType) *DigestDescriptor {
	return core2.NewDigestDescriptor(digest, typ)
}

func DefaultBlobDigesterRegistry() BlobDigesterRegistry {
	return core2.DefaultBlobDigesterRegistry
}

func DefaultContext() core2.Context {
	return core2.DefaultContext
}

func ForRepo(ctxtype, repostype string) BlobHandlerKey {
	return core2.ForRepo(ctxtype, repostype)
}

func ForMimeType(mimetype string) BlobHandlerKey {
	return core2.ForMimeType(mimetype)
}

func RegisterRepositorySpecHandler(handler RepositorySpecHandler, types ...string) {
	core2.RegisterRepositorySpecHandler(handler, types...)
}

func RegisterBlobHandler(handler BlobHandler, keys ...BlobHandlerKey) {
	core2.RegisterBlobHandler(handler, keys...)
}

func RegisterRepositoryType(name string, atype RepositoryType) {
	core2.DefaultRepositoryTypeScheme.Register(name, atype)
}

func RegisterAccessType(atype AccessType) {
	core2.DefaultAccessTypeScheme.Register(atype.GetKind(), atype)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core2.ToGenericRepositorySpec(spec)
}

const KIND_COMPONENTVERSION = core2.KIND_COMPONENTVERSION

func ErrUnknownComponentVersion(name, version string) error {
	return core2.ErrUnknownComponentVersion(name, version)
}
