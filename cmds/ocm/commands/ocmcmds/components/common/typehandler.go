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

package common

import (
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm"
)

type Object struct {
	Spec             ocm.RefSpec
	ComponentVersion ocm.ComponentVersionAccess
}

func (o *Object) AsManifest() interface{} {
	return o.ComponentVersion.GetDescriptor()
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx     ocm.Context
	session  ocm.Session
	repobase ocm.Repository
}

func NewTypeHandler(octx ocm.Context, session ocm.Session, repobase ocm.Repository) utils.TypeHandler {
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
	lister := h.repobase.ComponentLister()
	if lister == nil {
		return nil, nil
	}
	list, err := lister.GetComponents("", true)
	if err != nil {
		return nil, err
	}
	var result []output.Object
	for _, l := range list {
		part, err := h.Get(l)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
		result = append(result, part...)
	}
	return result, nil
}

func (h *TypeHandler) Get(name string) ([]output.Object, error) {
	var component ocm.ComponentAccess
	var result []output.Object
	var err error

	spec := ocm.RefSpec{}
	repo := h.repobase
	if repo == nil {
		parsed, ns, v, err := h.session.EvaluateRef(h.octx, name)
		if err != nil {
			return nil, errors.Wrapf(err, "component version reference %q", name)
		}
		spec = *parsed
		component = ns
		if v != nil {
			result = append(result, &Object{
				Spec:             spec,
				ComponentVersion: v,
			})
			return result, nil
		}
	} else {
		comp := ocm.CompSpec{Component: ""}
		if name != "" {
			comp, err = ocm.ParseComp(name)
			if err != nil {
				return nil, errors.Wrapf(err, "reference %q", name)
			}
		}
		component, err = h.session.LookupComponent(repo, comp.Component)
		if err != nil {
			return nil, errors.Wrapf(err, "reference %q", name)
		}
		spec.UniformRepositorySpec = repo.GetSpecification().AsUniformSpec(h.octx)
		spec.Component = comp.Component
		spec.Version = comp.Version
	}

	if spec.IsVersion() {
		v, err := component.LookupVersion(spec.Reference())
		if err != nil {
			return nil, err
		}
		result = append(result, &Object{
			Spec:             spec,
			ComponentVersion: v,
		})
	} else {
		versions, err := component.ListVersions()
		if err != nil {
			return nil, err
		}
		for _, vers := range versions {
			v, err := component.LookupVersion(vers)
			if err != nil {
				return nil, err
			}
			t := vers
			s := spec
			s.Version = &t
			result = append(result, &Object{
				Spec:             s,
				ComponentVersion: v,
			})
		}
	}
	return result, nil
}
