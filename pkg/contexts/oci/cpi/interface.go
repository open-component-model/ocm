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
	"github.com/open-component-model/ocm/pkg/common/accessio"
	core2 "github.com/open-component-model/ocm/pkg/contexts/oci/core"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const CONTEXT_TYPE = core2.CONTEXT_TYPE

const CommonTransportFormat = core2.CommonTransportFormat

type Context = core2.Context
type Repository = core2.Repository
type RepositorySpecHandlers = core2.RepositorySpecHandlers
type RepositorySpecHandler = core2.RepositorySpecHandler
type UniformRepositorySpec = core2.UniformRepositorySpec
type RepositoryType = core2.RepositoryType
type RepositorySpec = core2.RepositorySpec
type GenericRepositorySpec = core2.GenericRepositorySpec
type ArtefactAccess = core2.ArtefactAccess
type Artefact = core2.Artefact
type ArtefactSource = core2.ArtefactSource
type ArtefactSink = core2.ArtefactSink
type BlobSource = core2.BlobSource
type BlobSink = core2.BlobSink
type NamespaceLister = core2.NamespaceLister
type NamespaceAccess = core2.NamespaceAccess
type ManifestAccess = core2.ManifestAccess
type IndexAccess = core2.IndexAccess
type BlobAccess = core2.BlobAccess
type DataAccess = core2.DataAccess
type RepositorySource = core2.RepositorySource

type Descriptor = ociv1.Descriptor

var DefaultContext = core2.DefaultContext

func RegisterRepositoryType(name string, atype RepositoryType) {
	core2.DefaultRepositoryTypeScheme.Register(name, atype)
}

func RegisterRepositorySpecHandler(handler RepositorySpecHandler, types ...string) {
	core2.RegisterRepositorySpecHandler(handler, types...)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core2.ToGenericRepositorySpec(spec)
}

const KIND_OCIARTEFACT = core2.KIND_OCIARTEFACT
const KIND_MEDIATYPE = accessio.KIND_MEDIATYPE
const KIND_BLOB = accessio.KIND_BLOB

func ErrUnknownArtefact(name, version string) error {
	return core2.ErrUnknownArtefact(name, version)
}

func ErrBlobNotFound(digest digest.Digest) error {
	return accessio.ErrBlobNotFound(digest)
}

func IsErrBlobNotFound(err error) bool {
	return accessio.IsErrBlobNotFound(err)
}
