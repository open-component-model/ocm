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
	"github.com/open-component-model/ocm/pkg/oci"
)

const T_OCINAMESPACE = "oci namespace"

type oci_namespace struct {
	base
	kind string
	oci.NamespaceAccess
}

func (r *oci_namespace) Type() string {
	if r.kind != "" {
		return r.kind
	}
	return T_OCINAMESPACE
}

func (r *oci_namespace) Set() {
	r.Builder.oci_nsacc = r.NamespaceAccess
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Namespace(name string, f ...func()) {
	b.expect(b.oci_repo, T_OCIREPOSITORY)
	r, err := b.oci_repo.LookupNamespace(name)
	b.failOn(err)
	b.configure(&oci_namespace{NamespaceAccess: r, kind: T_OCIARTEFACTSET}, f)
}
