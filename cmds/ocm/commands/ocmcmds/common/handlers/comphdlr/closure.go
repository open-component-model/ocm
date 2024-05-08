package comphdlr

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
)

////////////////////////////////////////////////////////////////////////////////

func ClosureExplode(opts *output.Options, e interface{}) []interface{} {
	return traverse(common.History{}, e.(*Object), opts.Context, lookupoption.From(opts))
}

func traverse(hist common.History, o *Object, octx out.Context, lookup ocm.ComponentVersionResolver) []interface{} {
	key := common.VersionedElementKey(o.ComponentVersion)
	if err := hist.Add(ocm.KIND_COMPONENTVERSION, key); err != nil {
		return nil
	}
	result := []interface{}{o}
	refs := o.ComponentVersion.GetDescriptor().References
	/*
		refs=append(refs[:0:0], refs...)
		sort.Sort(refs)
	*/
	found := map[common.NameVersion]bool{}
	for _, ref := range refs {
		key := ocm.ComponentRefKey(&ref)
		if found[key] {
			continue // skip same ref wit different attributes for recursion
		}
		found[key] = true
		vers := ref.Version
		nested, err := o.Repository.LookupComponentVersion(ref.ComponentName, vers)
		if err != nil {
			out.Errf(octx, "Warning: lookup nested component version %q:%s [%s]: %s\n", ref.ComponentName, vers, hist, err)
		}
		if nested == nil && lookup != nil {
			nested, err = lookup.LookupComponentVersion(ref.ComponentName, vers)
			if err != nil {
				if !errors.IsErrNotFound(err) {
					out.Errf(octx, "Warning: fallback lookup nested component version \"%s:%s\" [%s]: %s\n", ref.ComponentName, vers, hist, err)
				} else {
					err = nil
				}
			}
		}
		if err != nil {
			continue
		}
		obj := &Object{
			History:  hist.Copy(),
			Identity: ref.GetIdentity(refs),
			Spec: ocm.RefSpec{
				UniformRepositorySpec: o.Spec.UniformRepositorySpec,
				CompSpec: ocm.CompSpec{
					Component: ref.ComponentName,
					Version:   &vers,
				},
			},
			Repository: o.Repository,
			// Component:        comp,
			ComponentVersion: nested,
		}
		if nested == nil {
			result = append(result, obj)
		} else {
			result = append(result, traverse(hist, obj, octx, lookup)...)
		}
	}
	return result
}
