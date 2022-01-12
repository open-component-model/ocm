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

package ocm

import (
	"io"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
)

type ComponentRepository interface {
	GetSpecification() RepositorySpec
	GetComponent(name string, version string) (ComponentAccess, error)
	WriteComponent(ComponentAccess) (ComponentAccess, error)
}

type DataAccess interface {
	Get() ([]byte, error)
	Reader() (io.Reader, error)
}

type ComponentAccess interface {
	GetRepository() ComponentRepository

	GetName() string
	GetVersion() string

	GetDescriptor() (*compdesc.ComponentDescriptor, error)
	GetResource(meta *compdesc.Identity) (DataAccess, error)
	GetSource(meta *compdesc.Identity) (DataAccess, error)
}
