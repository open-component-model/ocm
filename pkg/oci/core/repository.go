// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/opencontainers/go-digest"
)

type Repository interface {
	ExistsArtefact(name string, version string) (bool, error)
	LookupArtefact(name string, version string) (ArtefactAccess, error)
	LookupNamespace(name string) (NamespaceAccess, error)
	Close() error
}

type RepositorySource interface {
	GetRepository() Repository
}

type BlobAccess = accessio.BlobAccess
type DataAccess = accessio.DataAccess

type NamespaceAccess interface {
	RepositorySource

	GetArtefactByTag(tag string) (ArtefactAccess, error)
	GetArtefact(digest.Digest) (ArtefactAccess, error)

	AddBlob(BlobAccess) error
	AddArtefact(Artefact) (BlobAccess, error)

	NewArtefact(...*artdesc.Artefact) (ArtefactAccess, error)
}

type Artefact interface {
	IsManifest() bool
	IsIndex() bool

	Artefact() *artdesc.Artefact
	Manifest() (*artdesc.Manifest, error)
	Index() (*artdesc.Index, error)
}

type ArtefactAccess interface {
	Artefact

	GetDescriptor() *artdesc.Artefact
	GetManifest(digest digest.Digest) (ManifestAccess, error)
	GetBlob(digest digest.Digest) (BlobAccess, error)

	AddBlob(BlobAccess) error
	AddArtefact(Artefact, *artdesc.Platform) (BlobAccess, error)
	AddLayer(BlobAccess, *artdesc.Descriptor) (int, error)
}

type ManifestAccess interface {
	Artefact

	GetDescriptor() *artdesc.Manifest
	GetBlobDescriptor(digest digest.Digest) *artdesc.Descriptor
	GetBlob(digest digest.Digest) (BlobAccess, error)

	AddBlob(BlobAccess) error
	AddLayer(BlobAccess, *artdesc.Descriptor) (int, error)
}

type IndexAccess interface {
	Artefact

	GetDescriptor() *artdesc.Index
	GetBlobDescriptor(digest digest.Digest) *artdesc.Descriptor
	GetBlob(digest digest.Digest) (BlobAccess, error)

	GetArtefact(digest digest.Digest) (ArtefactAccess, error)
	GetIndex(digest digest.Digest) (IndexAccess, error)
	GetManifest(digest digest.Digest) (ManifestAccess, error)

	AddBlob(BlobAccess) error
	AddArtefact(Artefact, *artdesc.Platform) (BlobAccess, error)
}
