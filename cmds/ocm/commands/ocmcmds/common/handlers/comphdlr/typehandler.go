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

package comphdlr

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/tree"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
)

func Elem(e interface{}) ocm.ComponentVersionAccess {
	return e.(*Object).ComponentVersion
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	History  common.History
	Identity metav1.Identity

	Spec             ocm.RefSpec
	Repository       ocm.Repository
	Component        ocm.ComponentAccess
	ComponentVersion ocm.ComponentVersionAccess
}

func (o *Object) AsManifest() interface{} {
	return o.ComponentVersion.GetDescriptor()
}

func (o *Object) GetHistory() common.History {
	return o.History
}

func (o *Object) IsNode() *common.NameVersion {
	nv := common.VersionedElementKey(o.ComponentVersion)
	return &nv
}

var _ common.HistorySource = (*Object)(nil)
var _ tree.Object = (*Object)(nil)

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx     clictx.OCM
	session  ocm.Session
	repobase ocm.Repository
}

func NewTypeHandler(octx clictx.OCM, session ocm.Session, repobase ocm.Repository) utils.TypeHandler {
	return &TypeHandler{
		octx:     octx,
		session:  session,
		repobase: repobase,
	}
}

func (h *TypeHandler) Close() error {
	return nil
}

func (h *TypeHandler) All() ([]output.Object, error) {
	if h.repobase == nil {
		return nil, nil
	}
	return h.all(h.repobase)
}

func (h *TypeHandler) all(repo ocm.Repository) ([]output.Object, error) {
	lister := repo.ComponentLister()
	if lister == nil {
		return nil, nil
	}
	list, err := lister.GetComponents("", true)
	if err != nil {
		return nil, err
	}
	var result []output.Object
	for _, l := range list {
		part, err := h.get(repo, utils.StringSpec(l))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
		result = append(result, part...)
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	return h.get(h.repobase, elemspec)
}

func (h *TypeHandler) get(repo ocm.Repository, elemspec utils.ElemSpec) ([]output.Object, error) {
	var component ocm.ComponentAccess
	var result []output.Object
	var err error

	name := elemspec.String()
	spec := ocm.RefSpec{}
	if repo == nil {
		evaluated, err := h.session.EvaluateComponentRef(h.octx.Context(), name)
		if err != nil {
			return nil, errors.Wrapf(err, "%s: invalid component version reference", name)
		}
		if evaluated.Version != nil {
			result = append(result, &Object{
				Spec:             evaluated.Ref,
				Repository:       evaluated.Repository,
				Component:        evaluated.Component,
				ComponentVersion: evaluated.Version,
			})
			return result, nil
		}
		spec = evaluated.Ref
		component = evaluated.Component
		repo = evaluated.Repository
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
		spec.UniformRepositorySpec = repo.GetSpecification().AsUniformSpec(h.octx.Context())
		spec.Component = comp.Component
		spec.Version = comp.Version
	}

	if spec.IsVersion() {
		v, err := component.LookupVersion(*spec.Version)
		if err != nil {
			return nil, err
		}
		result = append(result, &Object{
			Repository:       repo,
			Spec:             spec,
			Component:        component,
			ComponentVersion: v,
		})
	} else {
		if component == nil {
			return h.all(repo)
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
					Repository:       repo,
					Spec:             s,
					Component:        component,
					ComponentVersion: v,
				})
			}
		}
	}
	return result, nil
}
