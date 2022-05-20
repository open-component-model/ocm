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

package ocm

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type BlobSink interface {
	AddBlob(blob accessio.BlobAccess) error
}

// StorageContext is the context information passed for Blobhandler
// registered for context type oci.CONTEXT_TYPE.
type StorageContext interface {
	cpi.StorageContext
	BlobSink
}

type DefaultStorageContext struct {
	Version cpi.ComponentVersionAccess
	Sink    BlobSink
}

func New(vers cpi.ComponentVersionAccess, access BlobSink) StorageContext {
	return &DefaultStorageContext{
		Version: vers,
		Sink:    access,
	}
}

func (c *DefaultStorageContext) TargetComponentVersion() cpi.ComponentVersionAccess {
	return c.Version
}

func (c *DefaultStorageContext) AddBlob(blob accessio.BlobAccess) error {
	return c.Sink.AddBlob(blob)
}
