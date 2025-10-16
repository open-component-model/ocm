package comphdlr

import (
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/common/output"
)

////////////////////////////////////////////////////////////////////////////////

func ClosureExplode(opts *output.Options, e interface{}) []interface{} {
	return traverse(common.History{}, e.(*Object), opts.Context, opts.Session, lookupoption.From(opts))
}

func traverse(hist common.History, o *Object, octx out.Context, sess ocm.Session, lookup ocm.ComponentVersionResolver) []interface{} {
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
			continue // skip same ref with different attributes for recursion
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
			if sess != nil {
				sess.AddCloser(nested)
			}
			result = append(result, traverse(hist, obj, octx, sess, lookup)...)
		}
	}
	return result
}
