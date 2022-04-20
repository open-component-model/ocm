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
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output/out"
	"github.com/open-component-model/ocm/pkg/common"
	oci2 "github.com/open-component-model/ocm/pkg/contexts/oci"
)

////////////////////////////////////////////////////////////////////////////////

func ClosureExplode(opts *output.Options, e interface{}) []interface{} {
	return traverse(common.History{}, e.(*Object), opts.Context)
}

func traverse(hist common.History, o *Object, octx out.Context) []output.Object {
	blob, _ := o.Artefact.Blob()
	key := common.NewNameVersion("", blob.Digest().String())
	if err := hist.Add(oci2.KIND_OCIARTEFACT, key); err != nil {
		return nil
	}
	result := []output.Object{o}
	if o.Artefact.IsIndex() {
		refs := o.Artefact.IndexAccess().GetDescriptor().Manifests

		found := map[common.NameVersion]bool{}
		for _, ref := range refs {
			key := common.NewNameVersion("", ref.Digest.String())
			if found[key] {
				continue // skip same ref wit different attributes for recursion
			}
			found[key] = true
			nested, err := o.Namespace.GetArtefact(key.GetVersion())
			if err != nil {
				out.Errf(octx, "Warning: lookup nested artefact %q [%s]: %s\n", ref.Digest, hist, err)
			}
			var obj = &Object{
				History: hist.Copy(),
				Key:     Key(nested),
				Spec: oci2.RefSpec{
					UniformRepositorySpec: o.Spec.UniformRepositorySpec,
					Repository:            o.Spec.Repository,
					Digest:                &ref.Digest,
				},
				Namespace: o.Namespace,
				Artefact:  nested,
			}
			result = append(result, traverse(hist, obj, octx)...)
		}
	}
	output.Print(result, "traverse %s", blob.Digest())
	return result
}
