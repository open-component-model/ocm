package elemhdlr

import (
	"encoding/json"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/tree"
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
