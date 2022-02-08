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

package comparch

import (
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/cpi"
)

// BlobContainer is the interface for an element capable to store blobs
type BlobContainer interface {
	GetBlobData(name string) (cpi.DataAccess, error)

	// AddBlob stores a local blob together with the component and
	// potentially provides a global reference according to the OCI distribution spec
	// if the blob described an oci artefact.
	// The resultimg access information (global and local) is provided as
	// an access method specification usable in a component descriptor
	AddBlob(blob cpi.BlobAccess, refName string) (cpi.AccessSpec, error)
}

// ComponentContainer is an interface for an element to store component versions
type ComponentContainer interface {
	GetContext() cpi.Context
	IsReadOnly() bool
	IsClosed() bool

	AddComponentVersion(ComponentVersionAccess) error
	GetComponentVersion(version string) (cpi.ComponentVersionAccess, error)
	BlobContainer
}

// ComponentVersionContainer is the interface of an element hosting a component version
type ComponentVersionContainer interface {
	GetContext() cpi.Context
	IsReadOnly() bool
	IsClosed() bool
	Update() error

	GetDescriptor() *compdesc.ComponentDescriptor

	BlobContainer
}
