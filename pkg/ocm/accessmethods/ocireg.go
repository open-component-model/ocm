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

package accessmethods

import (
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

// OCIRegistryType is the access type of a oci registry.
const OCIRegistryType = "ociRegistry"

func init() {
	ocm.RegisterAccessType(NewAccessType(OCIRegistryType, &OCIRegistryAccessSpec{}))
}

// OCIRegistryAccessSpec describes the access for a oci registry.
type OCIRegistryAccessSpec struct {
	runtime.ObjectType `json:",inline"`

	// ImageReference is the actual reference to the oci image repository and tag.
	ImageReference string `json:"imageReference"`
}

func (a *OCIRegistryAccessSpec) GetName() string {
	return OCIRegistryType
}

func (a *OCIRegistryAccessSpec) AccessMethod(c ocm.ComponentAccess) (ocm.AccessMethod, error) {
	return newOCIRegistryAccessMethod(a)
}

////////////////////////////////////////////////////////////////////////////////

type OCIRegistryAccessMethod struct {
  spec *OCIRegistryAccessSpec
}

var _ ocm.AccessMethod = &OCIRegistryAccessMethod{}

func newOCIRegistryAccessMethod(a *OCIRegistryAccessSpec) (*OCIRegistryAccessMethod, error) {
	return &OCIRegistryAccessMethod{
		spec: a,
	}, nil
}

func (m *OCIRegistryAccessMethod) GetType() string {
	return OCIRegistryType
}