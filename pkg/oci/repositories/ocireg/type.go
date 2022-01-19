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

package ocireg

import (
	"github.com/gardener/ocm/pkg/errors"
	area "github.com/gardener/ocm/pkg/oci"
	areautils "github.com/gardener/ocm/pkg/oci/ociutils"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/runtime"
)

const (
	OCIRegistryRepositoryType   = "OCIRegistry"
	OCIRegistryRepositoryTypeV1 = OCIRegistryRepositoryType + "/v1"
)

func init() {
	area.RegisterRepositoryType(OCIRegistryRepositoryType, areautils.NewRepositoryType(OCIRegistryRepositoryType, &OCIRegistryRepositorySpec{}, localAccessChecker))
	area.RegisterRepositoryType(OCIRegistryRepositoryTypeV1, areautils.NewRepositoryType(OCIRegistryRepositoryTypeV1, &OCIRegistryRepositorySpec{}, localAccessChecker))
}

// OCIRegistryRepositorySpec describes an OCI registry interface backed by an oci registry.
type OCIRegistryRepositorySpec struct {
	runtime.ObjectTypeVersion `json:",inline"`
	// BaseURL is the base url of the repository to resolve artefacts.
	BaseURL string `json:"baseUrl"`
}

// NewOCIRegistryRepositorySpec creates a new OCIRegistryRepositorySpec
func NewOCIRegistryRepositorySpec(baseURL string) *OCIRegistryRepositorySpec {
	return &OCIRegistryRepositorySpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(OCIRegistryRepositoryType),
		BaseURL:           baseURL,
	}
}

func (a *OCIRegistryRepositorySpec) GetType() string {
	return OCIRegistryRepositoryType
}
func (a *OCIRegistryRepositorySpec) Repository(ctx area.Context) (area.Repository, error) {
	return nil, errors.ErrNotImplemented() // TODO
}

func localAccessChecker(ctx area.Context, a compdesc.AccessSpec) bool {
	name := a.GetName()
	return name == accessmethods.LocalBlobType
}
