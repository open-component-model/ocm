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
	"encoding/json"
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/runtime"
)

type Object struct {
	History common.History
	Version ocm.ComponentVersionAccess
	Spec    metav1.Identity
	Id      metav1.Identity
	Element compdesc.ElementMetaAccessor
}

func (o *Object) AsManifest() interface{} {
	return o.Element
}

func History(e interface{}) string {
	o := e.(*Object)
	if o.History == nil {
		return ""
	}
	return o.History.String()
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	repository ocm.Repository
	session    ocm.Session
	access     ocm.ComponentVersionAccess
	elemaccess func(ocm.ComponentVersionAccess) compdesc.ElementAccessor
	recursive  bool
}

func NewTypeHandler(repository ocm.Repository, session ocm.Session, access ocm.ComponentVersionAccess, recursive bool, elemaccess func(ocm.ComponentVersionAccess) compdesc.ElementAccessor) utils.TypeHandler {
	return &TypeHandler{
		repository: repository,
		session:    session,
		access:     access,
		elemaccess: elemaccess,
		recursive:  recursive,
	}
}

func (h *TypeHandler) Close() error {
	return nil
}

func (h *TypeHandler) All() ([]output.Object, error) {
	return h.execute(nil, h.access, func(access ocm.ComponentVersionAccess) ([]output.Object, error) { return h.all(access) })
}

func (h *TypeHandler) all(access ocm.ComponentVersionAccess) ([]output.Object, error) {
	result := []output.Object{}
	elemaccess := h.elemaccess(access)
	l := elemaccess.Len()
	for i := 0; i < l; i++ {
		e := elemaccess.Get(i)
		result = append(result, &Object{
			Version: h.access,
			Id:      e.GetMeta().GetIdentity(elemaccess),
			Element: e,
		})
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	return h.execute(nil, h.access, func(access ocm.ComponentVersionAccess) ([]output.Object, error) { return h.get(access, elemspec) })
}

func (h *TypeHandler) execute(hist common.History, access ocm.ComponentVersionAccess, f func(access ocm.ComponentVersionAccess) ([]output.Object, error)) ([]output.Object, error) {
	key := common.VersionedElementKey(access)
	if err := hist.Add(ocm.KIND_COMPONENTVERSION, key); err != nil {
		return nil, err
	}
	result, err := f(access)
	if h.recursive {
		for _, ref := range h.access.GetDescriptor().ComponentReferences {
			nested, err := h.repository.LookupComponentVersion(ref.ComponentName, ref.Version)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: lookup nested component %q [%s]: %s", ref.ComponentName, hist, err)
				continue
			}
			out, err := h.execute(hist, nested, f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: cannot handle %s: %s", hist, err)
				continue
			}
			result = append(result, out...)
		}
	}
	return result, err
}

func (h *TypeHandler) get(access ocm.ComponentVersionAccess, elemspec utils.ElemSpec) ([]output.Object, error) {
	var result []output.Object

	selector := elemspec.(metav1.Identity)
	elemaccess := h.elemaccess(access)
	l := elemaccess.Len()
	for i := 0; i < l; i++ {
		e := elemaccess.Get(i)
		m := e.GetMeta()
		eid := m.GetMatchBaseIdentity()
		ok, _ := selector.Match(eid)
		if ok {
			result = append(result, &Object{
				Version: h.access,
				Spec:    selector,
				Id:      m.GetIdentity(elemaccess),
				Element: e,
			})
		}
	}
	return result, nil
}

var MetaOutput = []string{"NAME", "VERSION", "IDENTITY"}

func MapMetaOutput(e interface{}) []string {
	p := e.(*Object)
	m := p.Element.GetMeta()
	return []string{m.Name, m.Version, p.Id.String()}
}

var AccessOutput = []string{"ACCESSTYPE", "ACCESSSPEC"}

func MapAccessOutput(e compdesc.AccessSpec) []string {
	a := ""
	data, err := json.Marshal(e)
	if err != nil {
		a = "invalid: " + err.Error()
	} else {
		var un map[string]interface{}
		err := json.Unmarshal(data, &un)
		if err != nil {
			a = "invalid: " + err.Error()
		} else {
			delete(un, runtime.ATTR_TYPE)
			data, _ = json.Marshal(un)
			a = string(data)
		}
	}
	return []string{e.GetKind(), a}
}
