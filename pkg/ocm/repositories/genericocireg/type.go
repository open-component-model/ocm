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
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	area "github.com/gardener/ocm/pkg/ocm/core"
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

type GenericOCIRepositoryBackendType struct {
	runtime.ObjectTypeVersion
	ocictx oci.Context
}

var _ area.RepositoryType = &GenericOCIRepositoryBackendType{}

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

func (t *GenericOCIRepositoryBackendType) LocalSupportForAccessSpec(ctx area.Context, a compdesc.AccessSpec) bool {
	name := a.GetName()
	return name == accessmethods.LocalBlobType
}

////////////////////////////////////////////////////////////////////////////////

type GenericOCIBackendSpec struct {
	oci.GenericRepositorySpecWrapper `json:",inline"`
	compreg.ComponentRepositoryMeta  `json:",inline"`
}

func NewGenericOCIBackendSpec(spec oci.RepositorySpec, meta *compreg.ComponentRepositoryMeta) *GenericOCIBackendSpec {
	if meta.ComponentNameMapping == "" {
		meta.ComponentNameMapping = compreg.OCIRegistryURLPathMapping
	}
	return &GenericOCIBackendSpec{
		GenericRepositorySpecWrapper: oci.WrapRepositorySpec(spec),
		ComponentRepositoryMeta:      *meta,
	}
}

func (s *GenericOCIBackendSpec) GetType() string {
	return s.RepositorySpec.GetType()
}

func (s *GenericOCIBackendSpec) SetType(typ string) {
	s.RepositorySpec.SetType(typ)
}


func (s *GenericOCIBackendSpec) GetName() string {
	return s.RepositorySpec.GetName()
}

func (s *GenericOCIBackendSpec) GetVersion() string {
	return s.RepositorySpec.GetVersion()
}

func (s *GenericOCIBackendSpec) Repository(ctx area.Context) (area.Repository, error) {
	r, err := s.RepositorySpec.Repository(ctx.OCIContext())
	if err != nil {
		return nil, err
	}
	return NewRepository(ctx, r)
}
