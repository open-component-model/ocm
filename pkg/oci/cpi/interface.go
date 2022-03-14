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
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const CONTEXT_TYPE = core.CONTEXT_TYPE

type Context = core.Context
type Repository = core.Repository
type RepositoryType = core.RepositoryType
type RepositorySpec = core.RepositorySpec
type GenericRepositorySpec = core.GenericRepositorySpec
type ArtefactAccess = core.ArtefactAccess
type Artefact = core.Artefact
type ArtefactSource = core.ArtefactSource
type ArtefactSink = core.ArtefactSink
type BlobSource = core.BlobSource
type BlobSink = core.BlobSink
type NamespaceAccess = core.NamespaceAccess
type ManifestAccess = core.ManifestAccess
type IndexAccess = core.IndexAccess
type BlobAccess = core.BlobAccess
type DataAccess = core.DataAccess
type RepositorySource = core.RepositorySource

type Descriptor = ociv1.Descriptor

var DefaultContext = core.DefaultContext

func RegisterRepositoryType(name string, atype RepositoryType) {
	core.DefaultRepositoryTypeScheme.Register(name, atype)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core.ToGenericRepositorySpec(spec)
}

const KIND_OCIARTEFACT = core.KIND_OCIARTEFACT
const KIND_MEDIATYPE = accessio.KIND_MEDIATYPE
const KIND_BLOB = accessio.KIND_BLOB

func ErrUnknownArtefact(name, version string) error {
	return core.ErrUnknownArtefact(name, version)
}

func ErrBlobNotFound(digest digest.Digest) error {
	return accessio.ErrBlobNotFound(digest)
}

func IsErrBlobNotFound(err error) bool {
	return accessio.IsErrBlobNotFound(err)
}
