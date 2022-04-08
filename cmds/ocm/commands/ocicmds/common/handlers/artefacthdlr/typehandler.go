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

package artefacthdlr

import (
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
)

func Elem(e interface{}) oci.ArtefactAccess {
	return e.(*Object).Artefact
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	Spec     oci.RefSpec
	Artefact oci.ArtefactAccess
}

type Manifest struct {
	Spec     oci.RefSpec
	Digest   string
	Manifest *artdesc.Artefact
}

func (o *Object) AsManifest() interface{} {
	var digest string
	b, err := o.Artefact.Blob()
	if err == nil {
		digest = b.Digest().String()
	} else {
		digest = err.Error()
	}
	return &Manifest{
		Spec:     o.Spec,
		Digest:   digest,
		Manifest: o.Artefact.GetDescriptor(),
	}
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx     clictx.OCI
	session  oci.Session
	repobase oci.Repository
}

func NewTypeHandler(octx clictx.OCI, session oci.Session, repobase oci.Repository) utils.TypeHandler {
	return &TypeHandler{
		octx:     octx,
		session:  session,
		repobase: repobase,
	}
}

func (h *TypeHandler) Close() error {
	return h.session.Close()
}

func (h *TypeHandler) All() ([]output.Object, error) {
	if h.repobase == nil {
		return nil, nil
	}
	lister := h.repobase.NamespaceLister()
	if lister == nil {
		return nil, nil
	}
	list, err := lister.GetNamespaces("", true)
	if err != nil {
		return nil, err
	}
	var result []output.Object
	for _, l := range list {
		part, err := h.Get(utils.StringSpec(l))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
		result = append(result, part...)
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	var namespace oci.NamespaceAccess
	var result []output.Object
	var err error

	name := elemspec.String()
	spec := oci.RefSpec{}
	repo := h.repobase
	if repo == nil {
		evaluated, err := h.session.EvaluateRef(h.octx.Context(), name, h.octx.GetAlias)
		if err != nil {
			return nil, errors.Wrapf(err, "repository %q", name)
		}
		spec = evaluated.Ref
		namespace = evaluated.Namespace
		if evaluated.Artefact != nil {
			result = append(result, &Object{
				Spec:     spec,
				Artefact: evaluated.Artefact,
			})
			return result, nil
		}
	} else {
		art := oci.ArtSpec{Repository: ""}
		if name != "" {
			art, err = oci.ParseArt(name)
			if err != nil {
				return nil, errors.Wrapf(err, "artefact reference %q", name)
			}
		}
		namespace, err = h.session.LookupNamespace(repo, art.Repository)
		if err != nil {
			return nil, errors.Wrapf(err, "reference %q", name)
		}
		spec.Host = repo.GetSpecification().Name()
		spec.Repository = art.Repository
		spec.Tag = art.Tag
		spec.Digest = art.Digest
	}

	if spec.IsVersion() {
		a, err := namespace.GetArtefact(spec.Version())
		if err != nil {
			return nil, err
		}
		result = append(result, &Object{
			Spec:     spec,
			Artefact: a,
		})
	} else {
		tags, err := namespace.ListTags()
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			a, err := namespace.GetArtefact(tag)
			if err != nil {
				return nil, err
			}
			t := tag
			s := spec
			s.Tag = &t
			result = append(result, &Object{
				Spec:     s,
				Artefact: a,
			})
		}
	}
	return result, nil
}
