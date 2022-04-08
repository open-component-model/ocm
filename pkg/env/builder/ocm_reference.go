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
	"github.com/gardener/ocm/pkg/ocm/compdesc"
)

type ocm_reference struct {
	base

	meta compdesc.ComponentReference
}

const T_OCMREF = "reference"

func (r *ocm_reference) Type() string {
	return T_OCMREF
}

func (r *ocm_reference) Set() {
	r.Builder.ocm_meta = &r.meta.ElementMeta
}

func (r *ocm_reference) Close() error {
	return r.ocm_vers.SetReference(&r.meta)
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Reference(name, comp, vers string, f ...func()) {
	b.expect(b.ocm_vers, T_OCMVERSION)
	r := &ocm_reference{}
	r.meta.Name = name
	r.meta.Version = vers
	r.meta.ComponentName = comp
	b.configure(r, f)
}
