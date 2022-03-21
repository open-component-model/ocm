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
	"github.com/gardener/ocm/pkg/ocm/core"
)

const CONTEXT_TYPE = core.CONTEXT_TYPE

const CommonTransportFormat = core.CommonTransportFormat

type Context = core.Context
type Repository = core.Repository
type RepositorySpecHandlers = core.RepositorySpecHandlers
type RepositorySpecHandler = core.RepositorySpecHandler
type UniformRepositorySpec = core.UniformRepositorySpec
type ComponentLister = core.ComponentLister
type ComponentAccess = core.ComponentAccess
type ComponentVersionAccess = core.ComponentVersionAccess
type AccessSpec = core.AccessSpec
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
type GenericRepositorySpec = core.GenericRepositorySpec
type RepositoryType = core.RepositoryType

type BlobHandler = core.BlobHandler
type BlobHandlerKey = core.BlobHandlerKey
type StorageContext = core.StorageContext

func ForRepo(ctxtype, repostype string) BlobHandlerKey {
	return core.ForRepo(ctxtype, repostype)
}

func ForMimeType(mimetype string) BlobHandlerKey {
	return core.ForMimeType(mimetype)
}

func RegisterRepositorySpecHandler(handler RepositorySpecHandler, types ...string) {
	core.RegisterRepositorySpecHandler(handler, types...)
}

func RegisterBlobHandler(handler BlobHandler, keys ...BlobHandlerKey) {
	core.RegisterBlobHandler(handler, keys...)
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

const KIND_COMPONENTVERSION = core.KIND_COMPONENTVERSION

func ErrUnknownComponentVersion(name, version string) error {
	return core.ErrUnknownComponentVersion(name, version)
}
