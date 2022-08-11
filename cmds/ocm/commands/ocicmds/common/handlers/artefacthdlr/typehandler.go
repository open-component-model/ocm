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

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/tree"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

func Elem(e interface{}) oci.ArtefactAccess {
	return e.(*Object).Artefact
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	History    common.History
	Key        common.NameVersion
	Spec       oci.RefSpec
	AttachKind string
	Namespace  oci.NamespaceAccess
	Artefact   oci.ArtefactAccess
}

var _ common.HistoryElement = (*Object)(nil)
var _ tree.Object = (*Object)(nil)
var _ tree.Typed = (*Object)(nil)

func (o *Object) GetHistory() common.History {
	return o.History
}

func (o *Object) GetKey() common.NameVersion {
	return o.Key
}

func (o *Object) GetKind() string {
	return o.AttachKind
}

func (o *Object) IsNode() *common.NameVersion {
	blob, _ := o.Artefact.Blob()
	nv := common.NewNameVersion("", blob.Digest().String())
	return &nv
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

func (o *Object) String() string {
	blob, _ := o.Artefact.Blob()
	dig := blob.Digest()
	tag := "-"
	if o.Spec.Tag != nil {
		tag = *o.Spec.Tag
	}
	return fmt.Sprintf("%s [%s]: %s", dig, tag, o.History)
}

type Manifest struct {
	Spec     oci.RefSpec
	Digest   string
	Manifest *artdesc.Artefact
}

////////////////////////////////////////////////////////////////////////////////

func Key(a oci.ArtefactAccess) common.NameVersion {
	blob, _ := a.Blob()
	return common.NewNameVersion("", blob.Digest().String())
}

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
		part, err := h.get(utils.StringSpec(l))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
		result = append(result, part...)
	}
	output.Print(result, "all")
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	result, err := h.get(elemspec)
	output.Print(result, "get %s", elemspec)
	return result, err
}

func (h *TypeHandler) get(elemspec utils.ElemSpec) ([]output.Object, error) {
	var namespace oci.NamespaceAccess
	var result []output.Object
	var err error

	name := elemspec.String()
	spec := oci.RefSpec{}
	repo := h.repobase
	if repo == nil {
		evaluated, err := h.session.EvaluateRef(h.octx.Context(), name)
		if err != nil {
			return nil, errors.Wrapf(err, "repository %q", name)
		}
		spec = evaluated.Ref
		namespace = evaluated.Namespace
		if evaluated.Artefact != nil {
			obj := &Object{
				Key:       Key(evaluated.Artefact),
				Spec:      spec,
				Namespace: namespace,
				Artefact:  evaluated.Artefact,
			}
			result = append(result, obj)
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
		h.session.AddCloser(a)
		obj := &Object{
			Key:       Key(a),
			Spec:      spec,
			Namespace: namespace,
			Artefact:  a,
		}
		result = append(result, obj)
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
			h.session.AddCloser(a)
			t := tag
			s := spec
			s.Tag = &t
			result = append(result, &Object{
				Key:       Key(a),
				Spec:      s,
				Namespace: namespace,
				Artefact:  a,
			})
		}
	}
	return result, nil
}
