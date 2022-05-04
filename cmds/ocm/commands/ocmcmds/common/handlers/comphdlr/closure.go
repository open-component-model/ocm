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
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/out"
)

////////////////////////////////////////////////////////////////////////////////

func ClosureExplode(opts *output.Options, e interface{}) []interface{} {
	return traverse(common.History{}, e.(*Object), opts.Context, lookupoption.From(opts))
}

func traverse(hist common.History, o *Object, octx out.Context, lookup *lookupoption.Option) []interface{} {
	key := common.VersionedElementKey(o.ComponentVersion)
	if err := hist.Add(ocm.KIND_COMPONENTVERSION, key); err != nil {
		return nil
	}
	result := []interface{}{o}
	refs := o.ComponentVersion.GetDescriptor().ComponentReferences
	/*
		refs=append(refs[:0:0], refs...)
		sort.Sort(refs)
	*/
	found := map[common.NameVersion]bool{}
	for _, ref := range refs {
		key := common.NewNameVersion(ref.ComponentName, ref.Version)
		if found[key] {
			continue // skip same ref wit different attributes for recursion
		}
		found[key] = true
		var nested ocm.ComponentVersionAccess
		vers := ref.Version
		comp, err := o.Repository.LookupComponent(ref.ComponentName)
		if err != nil {
			out.Errf(octx, "Warning: lookup nested component %q [%s]: %s\n", ref.ComponentName, hist, err)
		} else {
			nested, err = comp.LookupVersion(vers)
			if err != nil {
				out.Errf(octx, "Warning: lookup nested component %q [%s]: %s\n", ref.ComponentName, hist, err)
			}
		}
		if nested == nil {
			comp, nested, err = lookup.LookupComponentVersion(ref.ComponentName, vers)
			if err != nil {
				out.Errf(octx, "Warning: fallback lookup nested component version \"%s:%s\" [%s]: %s\n", ref.ComponentName, vers, hist, err)
				continue
			}
		}
		var obj = &Object{
			History:  hist,
			Identity: ref.GetIdentity(refs),
			Spec: ocm.RefSpec{
				UniformRepositorySpec: o.Spec.UniformRepositorySpec,
				CompSpec: ocm.CompSpec{
					Component: ref.ComponentName,
					Version:   &vers,
				},
			},
			Repository:       o.Repository,
			Component:        comp,
			ComponentVersion: nested,
		}
		result = append(result, traverse(hist, obj, octx, lookup)...)
	}
	return result
}
