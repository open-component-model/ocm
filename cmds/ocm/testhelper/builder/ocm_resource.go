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
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
)

type ocm_resource struct {
	base

	meta   compdesc.ResourceMeta
	access compdesc.AccessSpec
	blob   accessio.BlobAccess
}

const T_OCMRESOURCE = "resource"

func (r *ocm_resource) Type() string {
	return T_OCMRESOURCE
}

func (r *ocm_resource) Set() {
	r.Builder.ocm_rsc = &r.meta
	r.Builder.ocm_acc = &r.access
	r.Builder.ocm_meta = &r.meta.ElementMeta
	r.Builder.blob = &r.blob
}

func (r *ocm_resource) Close() error {
	switch {
	case r.access != nil:
		return r.Builder.ocm_vers.SetResource(&r.meta, r.access)
	case r.blob != nil:
		return r.Builder.ocm_vers.SetResourceBlob(&r.meta, r.blob, "", nil)
	default:
		return errors.New("access or blob required")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Resource(name, vers, typ string, relation metav1.ResourceRelation, f ...func()) {
	b.expect(b.ocm_vers, T_OCMVERSION)
	r := &ocm_resource{}
	r.meta.Name = name
	r.meta.Version = vers
	r.meta.Type = typ
	r.meta.Relation = relation
	b.configure(r, f)
}
