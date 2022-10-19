// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/tree"
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
	Print(data.Slice(s), "tree transform")
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
