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

package genericocireg

import (
	"encoding/json"
	"fmt"

	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	compreg "github.com/gardener/ocm/pkg/ocm/repositories/ocireg"
	"github.com/gardener/ocm/pkg/runtime"
)

// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
// to OCI Image References.
type ComponentNameMapping string

const (
	OCIRegistryURLPathMapping ComponentNameMapping = "urlPath"
	OCIRegistryDigestMapping  ComponentNameMapping = "sha256-digest"
)

func init() {
	cpi.RegisterOCIImplementation(func(ctx oci.Context) (cpi.RepositoryType, error) {
		return NewOCIRepositoryBackendType(ctx), nil
	})
}

type GenericOCIRepositoryBackendType struct {
	runtime.ObjectTypeVersion
	ocictx oci.Context
}

var _ cpi.RepositoryType = &GenericOCIRepositoryBackendType{}

// NewOCIRepositoryBackendType creates generic type for any OCI Repository Backend
func NewOCIRepositoryBackendType(ocictx oci.Context) *GenericOCIRepositoryBackendType {
	return &GenericOCIRepositoryBackendType{
		ObjectTypeVersion: runtime.NewObjectTypeVersion("genericOCIRepositoryBackend"),
		ocictx:            ocictx,
	}
}

func (t *GenericOCIRepositoryBackendType) Decode(data []byte, unmarshal runtime.Unmarshaler) (runtime.TypedObject, error) {
	ospec, err := t.ocictx.RepositoryTypes().DecodeRepositorySpec(data, unmarshal)
	if err != nil {
		return nil, err
	}

	meta := &compreg.ComponentRepositoryMeta{}
	if unmarshal == nil {
		unmarshal = runtime.DefaultYAMLEncoding.Unmarshaler
	}
	err = unmarshal.Unmarshal(data, meta)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal component repository meta information")
	}
	return NewGenericOCIBackendSpec(ospec, meta), nil
}

func (t *GenericOCIRepositoryBackendType) LocalSupportForAccessSpec(ctx cpi.Context, a compdesc.AccessSpec) bool {
	name := a.GetName()
	return name == accessmethods.LocalBlobType
}

////////////////////////////////////////////////////////////////////////////////

type GenericOCIBackendSpec struct {
	oci.RepositorySpec
	compreg.ComponentRepositoryMeta
}

func NewGenericOCIBackendSpec(spec oci.RepositorySpec, meta *compreg.ComponentRepositoryMeta) *GenericOCIBackendSpec {
	if meta.ComponentNameMapping == "" {
		meta.ComponentNameMapping = compreg.OCIRegistryURLPathMapping
	}
	return &GenericOCIBackendSpec{
		RepositorySpec:          spec,
		ComponentRepositoryMeta: *meta,
	}
}

func (u *GenericOCIBackendSpec) UnmarshalJSON(data []byte) error {
	fmt.Printf("unmarshal generic ocireg spec %s\n", string(data))
	ocispec := &oci.GenericRepositorySpec{}
	if err := json.Unmarshal(data, ocispec); err != nil {
		return err
	}
	compmeta := &compreg.ComponentRepositoryMeta{}
	if err := json.Unmarshal(data, ocispec); err != nil {
		return err
	}

	u.RepositorySpec = ocispec
	u.ComponentRepositoryMeta = *compmeta
	return nil
}

// MarshalJSON implements a custom json unmarshal method for a unstructured type.
func (u *GenericOCIBackendSpec) MarshalJSON() ([]byte, error) {
	ocispec, err := runtime.ToUnstructuredTypedObject(u.RepositorySpec)
	if err != nil {
		return nil, err
	}
	compmeta, err := runtime.ToUnstructuredObject(u.ComponentRepositoryMeta)
	if err != nil {
		return nil, err
	}
	return json.Marshal(compmeta.FlatMerge(ocispec.Object))
}

func (s *GenericOCIBackendSpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	r, err := s.RepositorySpec.Repository(ctx.OCIContext(), creds)
	if err != nil {
		return nil, err
	}
	return NewRepository(ctx, r)
}
