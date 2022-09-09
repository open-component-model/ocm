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
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci/core"
)

const CONTEXT_TYPE = core.CONTEXT_TYPE

const CommonTransportFormat = core.CommonTransportFormat

type (
	Context                          = core.Context
	Repository                       = core.Repository
	RepositorySpecHandlers           = core.RepositorySpecHandlers
	RepositorySpecHandler            = core.RepositorySpecHandler
	UniformRepositorySpec            = core.UniformRepositorySpec
	RepositoryType                   = core.RepositoryType
	RepositorySpec                   = core.RepositorySpec
	IntermediateRepositorySpecAspect = core.IntermediateRepositorySpecAspect
	GenericRepositorySpec            = core.GenericRepositorySpec
	ArtefactAccess                   = core.ArtefactAccess
	Artefact                         = core.Artefact
	ArtefactSource                   = core.ArtefactSource
	ArtefactSink                     = core.ArtefactSink
	BlobSource                       = core.BlobSource
	BlobSink                         = core.BlobSink
	NamespaceLister                  = core.NamespaceLister
	NamespaceAccess                  = core.NamespaceAccess
	ManifestAccess                   = core.ManifestAccess
	IndexAccess                      = core.IndexAccess
	BlobAccess                       = core.BlobAccess
	DataAccess                       = core.DataAccess
	RepositorySource                 = core.RepositorySource
)

type Descriptor = ociv1.Descriptor

var DefaultContext = core.DefaultContext

func New(m ...datacontext.BuilderMode) Context {
	return core.Builder{}.New(m...)
}

func RegisterRepositoryType(name string, atype RepositoryType) {
	core.DefaultRepositoryTypeScheme.Register(name, atype)
}

func RegisterRepositorySpecHandler(handler RepositorySpecHandler, types ...string) {
	core.RegisterRepositorySpecHandler(handler, types...)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core.ToGenericRepositorySpec(spec)
}

func UniformRepositorySpecForHostURL(typ string, host string) *UniformRepositorySpec {
	return core.UniformRepositorySpecForHostURL(typ, host)
}

const (
	KIND_OCIARTEFACT = core.KIND_OCIARTEFACT
	KIND_MEDIATYPE   = accessio.KIND_MEDIATYPE
	KIND_BLOB        = accessio.KIND_BLOB
)

func ErrUnknownArtefact(name, version string) error {
	return core.ErrUnknownArtefact(name, version)
}

func ErrBlobNotFound(digest digest.Digest) error {
	return accessio.ErrBlobNotFound(digest)
}

func IsErrBlobNotFound(err error) bool {
	return accessio.IsErrBlobNotFound(err)
}
