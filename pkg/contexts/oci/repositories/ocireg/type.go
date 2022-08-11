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

package ocireg

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/containerd/containerd/reference"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	LegacyType = "ociRegistry"
	Type       = "OCIRegistry"
	TypeV1     = Type + runtime.VersionSeparator + "v1"

	ShortType   = "oci"
	ShortTypeV1 = ShortType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(LegacyType, cpi.NewRepositoryType(LegacyType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(Type, cpi.NewRepositoryType(Type, &RepositorySpec{}))
	cpi.RegisterRepositoryType(TypeV1, cpi.NewRepositoryType(TypeV1, &RepositorySpec{}))
	cpi.RegisterRepositoryType(ShortType, cpi.NewRepositoryType(ShortType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(ShortTypeV1, cpi.NewRepositoryType(ShortTypeV1, &RepositorySpec{}))
}

// Is checks the kind
func Is(spec cpi.RepositorySpec) bool {
	return spec != nil && spec.GetKind() == Type || spec.GetKind() == LegacyType
}

func IsKind(k string) bool {
	return k == Type || k == LegacyType
}

// RepositorySpec describes an OCI registry interface backed by an oci registry.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	// BaseURL is the base url of the repository to resolve artefacts.
	BaseURL     string `json:"baseUrl"`
	LegacyTypes *bool  `json:"legacyTypes,omitempty"`
}

var _ cpi.RepositorySpec = (*RepositorySpec)(nil)

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec(baseURL string) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		BaseURL:             baseURL,
	}
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Name() string {
	return a.BaseURL
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	var u *url.URL
	info := &RepositoryInfo{}
	legacy := false
	ref, err := reference.Parse(a.BaseURL)
	if err == nil {
		u, err = url.Parse("https://" + ref.Locator)
		if err != nil {
			return nil, err
		}
		info.Locator = ref.Locator
		if ref.Object != "" {
			return nil, fmt.Errorf("invalid repository locator %q", a.BaseURL)
		}
	} else {
		u, err = url.Parse(a.BaseURL)
		if err != nil {
			return nil, err
		}
		info.Locator = u.Host
	}
	if a.LegacyTypes != nil {
		legacy = *a.LegacyTypes
	} else {
		idx := strings.Index(info.Locator, "/")
		host := info.Locator
		if idx > 0 {
			host = info.Locator[:idx]
		}
		if host == "docker.io" {
			legacy = true
		}
	}
	info.Scheme = u.Scheme
	info.Creds = creds
	info.Legacy = legacy

	return NewRepository(ctx, a, info)
}
