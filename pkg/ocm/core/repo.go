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
	"io"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
)

type Repository interface {
	oci.Repository
	GetSpecification() RepositorySpec
	LookupComponent(name string, version string) (ComponentAccess, error)
	WriteComponent(ComponentAccess) (ComponentAccess, error)
}

type DataAccess = common.DataAccess
type BlobAccess = common.BlobAccess

type ResourceMeta = compdesc.ResourceMeta

type ResourceAccess interface {
	ResourceMeta() ResourceMeta
	AccessMethod() AccessMethod
	BlobAccess
}

type SourceMeta = compdesc.SourceMeta

type SourceAccess interface {
	SourceMeta() SourceMeta
	AccessMethod() AccessMethod
	BlobAccess
}

type ComponentAccess interface {
	common.VersionedElement
	io.Closer

	// GetAccessType returns the storage type of the component, which is the type of the repository
	// it is taken from
	GetAccessType() string

	GetRepository() Repository

	GetDescriptor() (*compdesc.ComponentDescriptor, error)
	GetResource(meta *metav1.Identity) (ResourceAccess, error)
	GetSource(meta *metav1.Identity) (ResourceAccess, error)
}
