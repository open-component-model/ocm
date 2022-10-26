// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	T_OCIARTEFACT = "artefact"
	T_OCIINDEX    = "index"
	T_OCIMANIFEST = "manifest"
)

type ociArtefact struct {
	base
	kind    string
	artfunc func(a oci.ArtefactAccess) error
	ns      cpi.NamespaceAccess
	cpi.ArtefactAccess
	tags []string
}

func (r *ociArtefact) Type() string {
	return r.kind
}

func (r *ociArtefact) Set() {
	r.Builder.oci_nsacc = r.ns
	r.Builder.oci_artacc = r.ArtefactAccess
	r.Builder.oci_cleanuplayers = true
	r.Builder.oci_tags = &r.tags

	if r.ns != nil {
		r.Builder.oci_artfunc = r.addArtefact
	}
}

func (r *ociArtefact) Close() error {
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

func (r *ociArtefact) addArtefact(a oci.ArtefactAccess) error {
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
	b.failOn(err, 1)
	tags := []string{}
	if tag != "" {
		tags = append(tags, tag)
	}
	r := b.configure(&ociArtefact{ArtefactAccess: v, kind: k, tags: tags, ns: ns, artfunc: b.oci_artfunc}, f, 1)
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
