// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package empty

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	Type   = "Empty"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

const ATTR_REPOS = "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/empty"

func init() {
	cpi.RegisterRepositoryType(Type, cpi.NewRepositoryType(Type, &RepositorySpec{}))
	cpi.RegisterRepositoryType(TypeV1, cpi.NewRepositoryType(TypeV1, &RepositorySpec{}))
}

// RepositorySpec describes an OCI registry interface backed by an oci registry.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
}

// NewRepositorySpec creates a new RepositorySpec.
func NewRepositorySpec() *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
	}
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Name() string {
	return Type
}

func (a *RepositorySpec) UniformRepositorySpec() *cpi.UniformRepositorySpec {
	u := &cpi.UniformRepositorySpec{
		Type: Type,
	}
	return u
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepository).(*Repository), nil
}
