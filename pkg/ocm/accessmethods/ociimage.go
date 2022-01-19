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

package accessmethods

import (
	"io"
	"io/ioutil"

	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/core/accesstypes"
	"github.com/gardener/ocm/pkg/ocm/runtime"
	"github.com/opencontainers/go-digest"
)

// OCIImageType is the access type of a oci registry.
const OCIImageType = "ociImage"
const OCIImageTypeV1 = OCIImageType + "/v1"

func init() {
	core.RegisterAccessType(accesstypes.NewType(OCIImageType, &OCIImageAccessSpec{}))
	core.RegisterAccessType(accesstypes.NewType(OCIImageTypeV1, &OCIImageAccessSpec{}))
}

// OCIImageAccessSpec describes the access for a oci image.
type OCIImageAccessSpec struct {
	runtime.ObjectTypeVersion `json:",inline"`

	// ImageReference is the actual reference to the oci image repository and tag.
	ImageReference string `json:"imageReference"`
}

// NewOCIRegistryAccessSpecV1 creates a new oci registry access spec version v1
func NewOCIImageAccessSpecV1(ref string) *OCIImageAccessSpec {
	return &OCIImageAccessSpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(OCIRegistryType),
		ImageReference:    ref,
	}
}

func (_ *OCIImageAccessSpec) GetType() string {
	return OCIRegistryType
}

func (a *OCIImageAccessSpec) ValidFor(core.Repository) bool {
	return true
}

func (a *OCIImageAccessSpec) AccessMethod(c core.ComponentAccess) (core.AccessMethod, error) {
	return newOCIImageAccessMethod(a)
}

////////////////////////////////////////////////////////////////////////////////

type OCIImageAccessMethod struct {
	spec *OCIImageAccessSpec
}

var _ core.AccessMethod = &OCIImageAccessMethod{}

func newOCIImageAccessMethod(a *OCIImageAccessSpec) (*OCIImageAccessMethod, error) {
	return &OCIImageAccessMethod{
		spec: a,
	}, nil
}

func (m *OCIImageAccessMethod) GetName() string {
	return OCIRegistryType
}

func (m *OCIImageAccessMethod) Open() (io.ReadCloser, error) {
	panic("no implemented") // TODO
}

func (m *OCIImageAccessMethod) Digest() digest.Digest {
	panic("no implemented") // TODO
}

func (m *OCIImageAccessMethod) Size() int64 {
	panic("no implemented") // TODO
}

func (m *OCIImageAccessMethod) Get() ([]byte, error) {
	file, err := m.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

func (m *OCIImageAccessMethod) Reader() (io.ReadCloser, error) {
	return m.Open()
}

func (m *OCIImageAccessMethod) MimeType() string {
	return ""
}
