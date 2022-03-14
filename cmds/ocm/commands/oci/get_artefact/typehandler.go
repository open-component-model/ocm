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

package get_artefact

import (
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
)

type Object struct {
	spec     oci.RefSpec
	artefact oci.ArtefactAccess
}

type Manifest struct {
	Spec     oci.RefSpec
	Manifest *artdesc.Artefact
}

func (o *Object) AsManifest() interface{} {
	return &Manifest{
		Spec:     o.spec,
		Manifest: o.artefact.GetDescriptor(),
	}
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx     oci.Context
	session  oci.Session
	repobase oci.Repository
}

func (h *TypeHandler) Close() error {
	return h.session.Close()
}

func (h *TypeHandler) Get(name string) ([]output.Object, error) {
	var namespace oci.NamespaceAccess
	var result []output.Object

	spec := oci.RefSpec{}
	repo := h.repobase
	if repo == nil {
		parsed, ns, err := h.session.EvaluateRef(h.octx, name)
		if err != nil {
			return nil, err
		}
		spec = *parsed
		namespace = ns
	} else {
		art, err := oci.ParseArt(name)
		if err != nil {
			return nil, err
		}
		namespace, err = h.session.LookupNamespace(repo, art.Repository)
		if err != nil {
			return nil, err
		}
		spec.Host = repo.GetSpecification().Name()
		spec.Repository = art.Repository
		spec.Tag = art.Tag
		spec.Digest = art.Digest
	}

	if spec.IsVersion() {
		a, err := namespace.GetArtefact(spec.Reference())
		if err != nil {
			return nil, err
		}
		result = append(result, &Object{
			spec:     spec,
			artefact: a,
		})
	} else {
		tags, err := namespace.ListTags()
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			a, err := namespace.GetArtefact(spec.Reference())
			if err != nil {
				return nil, err
			}
			t := tag
			s := spec
			s.Tag = &t
			result = append(result, &Object{
				spec:     s,
				artefact: a,
			})
		}
	}
	return result, nil
}
