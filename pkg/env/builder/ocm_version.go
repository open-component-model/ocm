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

package builder

import (
	"github.com/open-component-model/ocm/pkg/errors"
	metav1 "github.com/open-component-model/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/ocm/cpi"
)

const T_OCMVERSION = "component version"

type ocm_version struct {
	base
	kind string
	cpi.ComponentVersionAccess
}

func (r *ocm_version) Type() string {
	if r.kind != "" {
		return r.kind
	}
	return T_OCMVERSION
}

func (r *ocm_version) Set() {
	r.Builder.ocm_vers = r.ComponentVersionAccess
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Version(name string, f ...func()) {
	b.expect(b.ocm_comp, T_OCMCOMPONENT)
	v, err := b.ocm_comp.LookupVersion(name)
	if err != nil {
		if errors.IsErrNotFound(err) {
			v, err = b.ocm_comp.NewVersion(name)
		}
	}
	b.failOn(err)
	v.GetDescriptor().Provider = metav1.ProviderType("ACME")
	b.configure(&ocm_version{ComponentVersionAccess: v}, f)
}

func (b *Builder) ComponentVersion(name, version string, f ...func()) {
	b.Component(name, func() {
		b.Version(version, f...)
	})
}
