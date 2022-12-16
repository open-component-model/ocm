// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/tree"
)

type TreeOutputOption interface {
	ApplyTreeOutputOption(*TreeOutputOptions)
}

type TreeNodeMappingFunc func(*tree.TreeObject) []string

func (f TreeNodeMappingFunc) ApplyTreeOutputOption(o *TreeOutputOptions) {
	o.nodeMapping = f
}

type TreeNodeTitleFunc func(*tree.TreeObject) string

func (f TreeNodeTitleFunc) ApplyTreeOutputOption(o *TreeOutputOptions) {
	o.nodeTitle = f
}

type TreeOutputOptions struct {
	nodeMapping TreeNodeMappingFunc
	nodeTitle   TreeNodeTitleFunc
}

func (o TreeOutputOptions) ApplyTreeOutputOption(opts *TreeOutputOptions) {
	if o.nodeMapping != nil {
		opts.nodeMapping = o.nodeMapping
	}
	if o.nodeTitle != nil {
		opts.nodeTitle = o.nodeTitle
	}
}

func (o TreeOutputOptions) Apply(opts ...TreeOutputOption) TreeOutputOptions {
	for _, e := range opts {
		e.ApplyTreeOutputOption(&o)
	}
	return o
}

func (o TreeOutputOptions) NodeMapping(n int, obj *tree.TreeObject) interface{} {
	if o.nodeMapping == nil {
		return make([]string, n)
	}
	return o.nodeMapping(obj)
}

func (o TreeOutputOptions) NodeTitle(obj *tree.TreeObject) string {
	if o.nodeTitle == nil {
		return obj.Node.String()
	}
	return o.nodeTitle(obj)
}

func TreeOutput(t *TableOutput, header string, o ...TreeOutputOption) *TableOutput {
	opts := *t.Options
	opts.FixedColums = 1
	return &TableOutput{
		Headers: Fields(header, t.Headers),
		Options: &opts,
		Chain:   processing.Append(t.Chain, processing.Transform(treeTransform)),
		Mapping: treeMapping(len(t.Headers), t.Mapping, TreeOutputOptions{}.Apply(o...)),
	}
}

func treeTransform(s data.Iterable) data.Iterable {
	Print(data.Slice(s), "tree transform")
	result := tree.MapToTree(tree.ObjectSlice(s), nil)
	return result
}

func treeMapping(n int, m processing.MappingFunction, opts TreeOutputOptions) processing.MappingFunction {
	return func(e interface{}) interface{} {
		o := e.(*tree.TreeObject)
		if o.Object != nil {
			return Fields(o.Graph, m(o.Object))
		}
		return Fields(o.Graph+" "+opts.NodeTitle(o), opts.NodeMapping(n, o)) // create empty table line
	}
}
