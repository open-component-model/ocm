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
	GenericAccessSpec                = core.GenericAccessSpec
	AccessMethod                     = core.AccessMethod
	AccessMethodSupport              = core.AccessMethodSupport
	AccessType                       = core.AccessType
	DataAccess                       = core.DataAccess
	BlobAccess                       = core.BlobAccess
	SourceAccess                     = core.SourceAccess
	SourceMeta                       = core.SourceMeta
	ResourceAccess                   = core.ResourceAccess
	ResourceMeta                     = core.ResourceMeta
	RepositorySpec                   = core.RepositorySpec
	IntermediateRepositorySpecAspect = core.IntermediateRepositorySpecAspect
	GenericRepositorySpec            = core.GenericRepositorySpec
	RepositoryType                   = core.RepositoryType
	ComponentReference               = core.ComponentReference
)

type (
	BlobHandler                  = core.BlobHandler
	BlobHandlerOption            = core.BlobHandlerOption
	StorageContext               = core.StorageContext
	ImplementationRepositoryType = core.ImplementationRepositoryType
)

type (
	DigesterType         = core.DigesterType
	BlobDigester         = core.BlobDigester
	BlobDigesterRegistry = core.BlobDigesterRegistry
	DigestDescriptor     = core.DigestDescriptor
)

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

const (
	KIND_COMPONENTVERSION = core.KIND_COMPONENTVERSION
	KIND_RESOURCE         = core.KIND_RESOURCE
	KIND_SOURCE           = core.KIND_SOURCE
	KIND_REFERENCE        = core.KIND_REFERENCE
)

func ErrComponentVersionNotFound(name, version string) error {
	return core.ErrComponentVersionNotFound(name, version)
}

func ErrComponentVersionNotFoundWrap(err error, name, version string) error {
	return core.ErrComponentVersionNotFoundWrap(err, name, version)
}

// PrefixProvider is supported by RepositorySpecs to
// provide info about a potential path prefix to
// use for globalized local artifacts.
type PrefixProvider interface {
	PathPrefix() string
}

func RepositoryPrefix(spec RepositorySpec) string {
	if s, ok := spec.(PrefixProvider); ok {
		return s.PathPrefix()
	}
	return ""
}

// HintProvider is able to provide a name hint for globalization of local
// artifacts.
type HintProvider core.HintProvider

func ArtefactNameHint(spec AccessSpec, cv ComponentVersionAccess) string {
	if h, ok := spec.(HintProvider); ok {
		return h.GetReferenceHint(cv)
	}
	return ""
}
