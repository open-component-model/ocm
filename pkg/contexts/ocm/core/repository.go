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

package core

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

type ComponentVersionResolver interface {
	LookupComponentVersion(name string, version string) (ComponentVersionAccess, error)
}

type Repository interface {
	GetContext() Context

	GetSpecification() RepositorySpec
	ComponentLister() ComponentLister

	ExistsComponentVersion(name string, version string) (bool, error)
	LookupComponentVersion(name string, version string) (ComponentVersionAccess, error)
	LookupComponent(name string) (ComponentAccess, error)

	Close() error
}

type DataAccess = accessio.DataAccess
type BlobAccess = accessio.BlobAccess
type MimeType = accessio.MimeType

type ComponentAccess interface {
	GetContext() Context
	GetName() string

	ListVersions() ([]string, error)
	LookupVersion(version string) (ComponentVersionAccess, error)
	AddVersion(ComponentVersionAccess) error
	NewVersion(version string, overrides ...bool) (ComponentVersionAccess, error)

	Close() error
}
type ResourceMeta = compdesc.ResourceMeta
type ComponentReference = compdesc.ComponentReference

type BaseAccess interface {
	Access() (AccessSpec, error)
	AccessMethod() (AccessMethod, error)
}

type ResourceAccess interface {
	Meta() *ResourceMeta
	BaseAccess
}

type SourceMeta = compdesc.SourceMeta

type SourceAccess interface {
	Meta() *SourceMeta
	BaseAccess
}

type ComponentVersionAccess interface {
	common.VersionedElement

	Repository() Repository

	GetContext() Context

	GetDescriptor() *compdesc.ComponentDescriptor

	GetResources() []ResourceAccess
	GetResource(meta metav1.Identity) (ResourceAccess, error)
	GetResourceByIndex(i int) (ResourceAccess, error)

	GetSources() []SourceAccess
	GetSource(meta metav1.Identity) (SourceAccess, error)
	GetSourceByIndex(i int) (SourceAccess, error)

	GetReference(meta metav1.Identity) (ComponentReference, error)
	GetReferenceByIndex(i int) (ComponentReference, error)

	// AccessMethod provides an access method implementation for
	// an access spec. This might be a repository local implementation
	// or a global one. It might be called by the AccessSpec method
	// to map itself to a local implementation or called directly.
	// If called it should forward the call to the AccessSpec
	// if and only if this specs NOT states to be local IsLocal()==false
	// If the spec states to be local, the repository is responsible for
	// providing a local implementation or return nil if this is
	// not supported by the actual repository type.
	AccessMethod(AccessSpec) (AccessMethod, error)

	// AddBlob adds a local blob and returns an appropriate local access spec
	AddBlob(blob BlobAccess, refName string, global AccessSpec) (AccessSpec, error)

	SetResourceBlob(meta *ResourceMeta, blob BlobAccess, refname string, global AccessSpec) error
	SetResource(*ResourceMeta, compdesc.AccessSpec) error
	// AdjustResourceAccess is used to modify the access spec. The old and new one MUST refer to the same content.
	AdjustResourceAccess(meta *ResourceMeta, acc compdesc.AccessSpec) error

	SetSourceBlob(meta *SourceMeta, blob BlobAccess, refname string, global AccessSpec) error
	SetSource(*SourceMeta, compdesc.AccessSpec) error

	SetReference(ref *ComponentReference) error

	DiscardChanges()
	io.Closer
}

// ComponentLister provides the optional repository list functionality of
// a repository
type ComponentLister interface {
	// NumComponents returns the number of components found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	NumComponents(prefix string) (int, error)

	// GetNamespaces returns the name of namespaces found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	GetComponents(prefix string, closure bool) ([]string, error)
}
