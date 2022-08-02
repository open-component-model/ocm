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
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

const T_OCIARTEFACT = "artefact"
const T_OCIINDEX = "index"
const T_OCIMANIFEST = "manifest"

type oci_artefact struct {
	base
	kind    string
	artfunc func(a oci.ArtefactAccess) error
	ns      cpi.NamespaceAccess
	cpi.ArtefactAccess
	tags []string
}

func (r *oci_artefact) Type() string {
	return r.kind
}

func (r *oci_artefact) Set() {
	r.Builder.oci_nsacc = r.ns
	r.Builder.oci_artacc = r.ArtefactAccess
	r.Builder.oci_cleanuplayers = true
	r.Builder.oci_tags = &r.tags

	if r.ns != nil {
		r.Builder.oci_artfunc = r.addArtefact
	}
}

func (r *oci_artefact) Close() error {
	err := r.ArtefactAccess.Close()
	if err != nil {
		return err
	}
	blob, err := r.Builder.oci_nsacc.AddArtefact(r.ArtefactAccess, r.tags...)
	if err == nil && r.artfunc != nil {
		err = r.artfunc(r.ArtefactAccess)
	}
	if err == nil {
		r.result = artdesc.DefaultBlobDescriptor(blob)
	}
	return err
}

func (r *oci_artefact) addArtefact(a oci.ArtefactAccess) error {
	_, err := r.ArtefactAccess.AddArtefact(a, nil)
	return err
}

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) artefact(tag string, ns cpi.NamespaceAccess, t func(access oci.ArtefactAccess) (string, error), f ...func()) *artdesc.Descriptor {
	var k string
	b.expect(b.oci_nsacc, T_OCINAMESPACE)
	v, err := b.oci_nsacc.GetArtefact(tag)
	if err != nil {
		if errors.IsErrNotFound(err) {
			if b, _ := artdesc.IsDigest(tag); !b {
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
	b.failOn(err, 2)
	tags := []string{}
	if tag != "" {
		tags = append(tags, tag)
	}
	r := b.configure(&oci_artefact{ArtefactAccess: v, kind: k, tags: tags, ns: ns, artfunc: b.oci_artfunc}, f, 1)
	if r == nil {
		return nil
	}
	return r.(*artdesc.Descriptor)
}

func (b *Builder) Index(tag string, f ...func()) *artdesc.Descriptor {
	return b.artefact(tag, b.oci_nsacc, func(a oci.ArtefactAccess) (string, error) {
		if a.IndexAccess() == nil {
			return "", errors.Newf("artefact is manifest")
		}
		return T_OCIINDEX, nil
	}, f...)
}

func (b *Builder) Manifest(tag string, f ...func()) *artdesc.Descriptor {
	return b.artefact(tag, nil, func(a oci.ArtefactAccess) (string, error) {
		if a.ManifestAccess() == nil {
			return "", errors.Newf("artefact is index")
		}
		return T_OCIMANIFEST, nil
	}, f...)
}
