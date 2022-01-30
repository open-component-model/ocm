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
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Context = core.Context
type Repository = core.Repository
type RepositoryType = core.RepositoryType
type RepositorySpec = core.RepositorySpec
type GenericRepositorySpec = core.GenericRepositorySpec
type ArtefactAccess = core.ArtefactAccess
type ArtefactComposer = core.ArtefactComposer
type ManifestAccess = core.ManifestAccess
type BlobAccess = core.BlobAccess
type DataAccess = core.DataAccess

type Descriptor = ociv1.Descriptor

type RepositorySource interface {
	GetRepository() Repository
}

var DefaultContext = core.DefaultContext

func RegisterRepositoryType(name string, atype RepositoryType) {
	core.DefaultRepositoryTypeScheme.Register(name, atype)
}

const KIND_OCIARTEFACT = core.KIND_OCIARTEFACT
const KIND_MEDIATYPE = core.KIND_MEDIATYPE
const KIND_BLOB = core.KIND_BLOB

func ErrUnknownArtefact(name, version string) error {
	return core.ErrUnknownArtefact(name, version)
}

func ErrBlobNotFound(digest digest.Digest) error {
	return core.ErrBlobNotFound(digest)
}

func IsErrBlobNotFound(err error) bool {
	return core.IsErrBlobNotFound(err)
}
