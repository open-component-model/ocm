// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package elemhdlr

import (
	"encoding/json"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/tree"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var MetaOutput = []string{"NAME", "VERSION", "IDENTITY"}

func MapMetaOutput(e interface{}) []string {
	p := e.(*Object)
	m := p.Element.GetMeta()
	id := p.Id.Copy()
	id.Remove(metav1.SystemIdentityName)
	return []string{m.Name, m.Version, id.String()}
}

func MapNodeOutput(e interface{}) []string {
	p := e.(*Object)
	id := p.VersionId.Copy()
	id.Remove(metav1.SystemIdentityName)
	return []string{p.VersionId[metav1.SystemIdentityName], p.Version.GetVersion(), id.String()}
}

var AccessOutput = []string{"ACCESSTYPE", "ACCESSSPEC"}

func MapAccessOutput(e compdesc.AccessSpec) []string {
	data, err := json.Marshal(e)
	if err != nil {
		return []string{e.GetKind(), err.Error()}
	}

	var un map[string]interface{}
	if err := json.Unmarshal(data, &un); err != nil {
		return []string{e.GetKind(), err.Error()}
	}

	delete(un, runtime.ATTR_TYPE)

	data, err = json.Marshal(un)
	if err != nil {
		return []string{e.GetKind(), err.Error()}
	}
	return []string{e.GetKind(), string(data)}
}

func NodeMapping(n int) output.TreeOutputOption {
	return output.TreeNodeMappingFunc(func(obj *tree.TreeObject) []string {
		r := MapNodeOutput(obj.Node.CausedBy)
		for n > 0 {
			r = append(r, "")
			n--
		}
		return r
	})
}

func NodeTitle(obj *tree.TreeObject) string {
	return obj.Node.GetName()
}
