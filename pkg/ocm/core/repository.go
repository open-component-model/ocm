// Copyright 2022 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
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
	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
)

type Repository interface {
	GetContext() Context

	GetSpecification() RepositorySpec
	ExistsComponentVersion(name string, version string) (bool, error)
	LookupComponentVersion(name string, version string) (ComponentVersionAccess, error)
	LookupComponent(name string) (ComponentAccess, error)
}

type DataAccess = accessio.DataAccess
type BlobAccess = accessio.BlobAccess
type MimeType = accessio.MimeType

type ComponentAccess interface {
	GetContext() Context
	GetName() string

	LookupVersion(version string) (ComponentVersionAccess, error)
	AddVersion(ComponentVersionAccess) error
	NewVersion(version string) (ComponentVersionAccess, error)
}
type ResourceMeta = compdesc.ResourceMeta

type BaseAccess interface {
	AccessMethod() (AccessMethod, error)
	DataAccess
}

type ResourceAccess interface {
	Meta() ResourceMeta
	BaseAccess
}

type SourceMeta = compdesc.SourceMeta

type SourceAccess interface {
	Meta() SourceMeta
	BaseAccess
}

type ComponentVersionAccess interface {
	common.VersionedElement

	GetContext() Context

	GetDescriptor() *compdesc.ComponentDescriptor
	GetResource(meta metav1.Identity) (ResourceAccess, error)
	GetSource(meta metav1.Identity) (SourceAccess, error)

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
	AddBlob(blob BlobAccess, refName string) (AccessSpec, error)

	AddResourceBlob(meta *ResourceMeta, blob BlobAccess, refname string) error
	AddResource(*ResourceMeta, compdesc.AccessSpec) error

	AddSourceBlob(meta *SourceMeta, blob BlobAccess, refname string) error
	AddSource(*SourceMeta, compdesc.AccessSpec) error
}
