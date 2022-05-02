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
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
)

type Repository interface {
	GetSpecification() RepositorySpec
	NamespaceLister() NamespaceLister

	ExistsArtefact(name string, ref string) (bool, error)
	LookupArtefact(name string, ref string) (ArtefactAccess, error)
	LookupNamespace(name string) (NamespaceAccess, error)
	Close() error
}

type RepositorySource interface {
	GetRepository() Repository
}

type BlobAccess = accessio.BlobAccess
type DataAccess = accessio.DataAccess

type BlobSource interface {
	GetBlobData(digest digest.Digest) (DataAccess, error)
}

type BlobSink interface {
	AddBlob(BlobAccess) error
}

type ArtefactSink interface {
	AddBlob(BlobAccess) error
	AddArtefact(a Artefact, tags ...string) (BlobAccess, error)
	AddTags(digest digest.Digest, tags ...string) error
}

type ArtefactSource interface {
	GetArtefact(version string) (ArtefactAccess, error)
	GetBlobData(digest digest.Digest) (DataAccess, error)
}

type NamespaceAccess interface {
	ArtefactSource
	ArtefactSink

	GetNamespace() string
	ListTags() ([]string, error)

	NewArtefact(...*artdesc.Artefact) (ArtefactAccess, error)

	Close() error
}

type Artefact interface {
	IsManifest() bool
	IsIndex() bool

	Digest() digest.Digest
	Blob() (BlobAccess, error)
	Artefact() *artdesc.Artefact
	Manifest() (*artdesc.Manifest, error)
	Index() (*artdesc.Index, error)
}

type ArtefactAccess interface {
	Artefact
	BlobSource
	BlobSink

	GetDescriptor() *artdesc.Artefact
	ManifestAccess() ManifestAccess
	IndexAccess() IndexAccess
	GetBlob(digest digest.Digest) (BlobAccess, error)

	AddBlob(BlobAccess) error

	AddArtefact(Artefact, *artdesc.Platform) (BlobAccess, error)
	AddLayer(BlobAccess, *artdesc.Descriptor) (int, error)

	Close() error
}

type ManifestAccess interface {
	Artefact

	GetDescriptor() *artdesc.Manifest
	GetBlobDescriptor(digest digest.Digest) *artdesc.Descriptor
	GetConfigBlob() (BlobAccess, error)
	GetBlob(digest digest.Digest) (BlobAccess, error)

	AddBlob(BlobAccess) error
	AddLayer(BlobAccess, *artdesc.Descriptor) (int, error)
	SetConfigBlob(blob BlobAccess, d *artdesc.Descriptor) error
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

// NamespaceLister provides the optional repository list functionality of
// a repository
type NamespaceLister interface {
	// NumNamespaces returns the number of namespaces found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	NumNamespaces(prefix string) (int, error)

	// GetNamespaces returns the name of namespaces found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	GetNamespaces(prefix string, closure bool) ([]string, error)
}
