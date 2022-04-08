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
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
)

const T_OCIARTEFACT = "artefact"
const T_OCIINDEX = "index"
const T_OCIMANIFEST = "manifest"

type oci_artefact struct {
	base
	kind string
	cpi.ArtefactAccess
	tags []string
}

func (r *oci_artefact) Type() string {
	return r.kind
}

func (r *oci_artefact) Set() {
	r.Builder.oci_artacc = r.ArtefactAccess
	r.Builder.oci_tags = &r.tags
}

func (r *oci_artefact) Close() error {
	err := r.ArtefactAccess.Close()
	if err != nil {
		return err
	}
	_, err = r.Builder.oci_nsacc.AddArtefact(r.ArtefactAccess, r.tags...)
	return err
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) artefact(version string, t func(access oci.ArtefactAccess) (string, error), f ...func()) {
	var k string
	b.expect(b.oci_nsacc, T_OCINAMESPACE)
	v, err := b.oci_nsacc.GetArtefact(version)
	if err != nil {
		if errors.IsErrNotFound(err) {
			if b, _ := artdesc.IsDigest(version); !b {
				err = nil
			}
		}
	}
	if v == nil {
		v, err = b.oci_nsacc.NewArtefact()
	}
	if v != nil {
		k, err = t(v)
	}
	b.failOn(err)
	tags := []string{}
	if version != "" {
		tags = append(tags, version)
	}
	b.configure(&oci_artefact{ArtefactAccess: v, kind: k, tags: tags}, f)
}

func (b *Builder) Index(version string, f ...func()) {
	b.artefact(version, func(a oci.ArtefactAccess) (string, error) {
		if a.IndexAccess() == nil {
			return "", errors.Newf("artefact is manifest")
		}
		return T_OCIINDEX, nil
	}, f...)
}

func (b *Builder) Manifest(version string, f ...func()) {
	b.artefact(version, func(a oci.ArtefactAccess) (string, error) {
		if a.ManifestAccess() == nil {
			return "", errors.Newf("artefact is index")
		}
		return T_OCIMANIFEST, nil
	}, f...)
}
