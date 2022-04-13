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

package output

import (
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"github.com/gardener/ocm/cmds/ocm/pkg/processing"
	"github.com/gardener/ocm/cmds/ocm/pkg/tree"
)

func TreeOutput(t *TableOutput, header string) *TableOutput {
	opts := *t.Options
	opts.FixedColums = 1
	return &TableOutput{
		Headers: Fields(header, t.Headers),
		Options: &opts,
		Chain:   processing.Append(t.Chain, processing.Transform(treeTransform)),
		Mapping: treeMapping(len(t.Headers), t.Mapping),
	}
}

func treeTransform(s data.Iterable) data.Iterable {
	result := tree.MapToTree(tree.ObjectSlice(s), nil)
	return result
}

func treeMapping(n int, m processing.MappingFunction) processing.MappingFunction {
	return func(e interface{}) interface{} {
		o := e.(*tree.TreeObject)
		if o.Object != nil {
			return Fields(o.Graph, m(o.Object))
		}
		return Fields(o.Graph+" "+o.Node.String(), make([]string, n)) // create empty table line
	}
}
