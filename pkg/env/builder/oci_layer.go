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
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

const T_OCILAYER = "oci layer"

type oci_layer struct {
	base
	blob accessio.BlobAccess
}

func (r *oci_layer) Type() string {
	return T_OCILAYER
}

func (r *oci_layer) Set() {
	r.Builder.blob = &r.blob
}

func (r *oci_layer) Close() error {
	if r.blob == nil {
		return errors.Newf("config blob required")
	}
	m := r.Builder.oci_artacc.ManifestAccess()
	m.AddLayer(r.blob, nil)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Layer(f ...func()) {
	b.expect(b.oci_artacc, T_OCIMANIFEST, func() bool { return b.oci_artacc.IsManifest() })
	b.configure(&oci_layer{}, f)
}
