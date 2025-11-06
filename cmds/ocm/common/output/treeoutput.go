package output

import (
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/cmds/ocm/common/data"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/tree"
)

type TreeOutputOption interface {
	ApplyTreeOutputOption(*TreeOutputOptions)
}

type TreeNodeMappingFunc func(*tree.TreeObject) []string

func (f TreeNodeMappingFunc) ApplyTreeOutputOption(o *TreeOutputOptions) {
	o.nodeMapping = f
}

type TreeSynthesizedTitleFunc func(*tree.TreeObject) string

func (f TreeSynthesizedTitleFunc) ApplyTreeOutputOption(o *TreeOutputOptions) {
	o.syntTitle = f
}

type TreeElemTitleFunc func(*tree.TreeObject) string

func (f TreeElemTitleFunc) ApplyTreeOutputOption(o *TreeOutputOptions) {
	o.elemTitle = f
}

type TreeSymbol string

func (s TreeSymbol) ApplyTreeOutputOption(o *TreeOutputOptions) {
	o.symbol = generics.PointerTo(string(s))
}

type TreeOutputOptions struct {
	nodeMapping TreeNodeMappingFunc
	syntTitle   TreeSynthesizedTitleFunc
	elemTitle   TreeElemTitleFunc
	symbol      *string
}

func (o TreeOutputOptions) ApplyTreeOutputOption(opts *TreeOutputOptions) {
	if o.nodeMapping != nil {
		opts.nodeMapping = o.nodeMapping
	}
	if o.syntTitle != nil {
		opts.syntTitle = o.syntTitle
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
	if o.syntTitle == nil {
		return obj.Node.String()
	}
	return o.syntTitle(obj)
}

func (o TreeOutputOptions) ElemTitle(obj *tree.TreeObject) string {
	if o.elemTitle == nil {
		return ""
	}
	return " " + o.elemTitle(obj)
}

func TreeOutput(t *TableOutput, header string, o ...TreeOutputOption) *TableOutput {
	opts := *t.Options
	opts.FixedColums = 1
	topts := TreeOutputOptions{}.Apply(o...)
	return &TableOutput{
		Headers: Fields(header, t.Headers),
		Options: &opts,
		Chain:   processing.Append(t.Chain, processing.Transform(transformer{topts.symbol}.treeTransform)),
		Mapping: treeMapping(len(t.Headers), t.Mapping, topts),
	}
}

type transformer struct {
	symbol *string
}

func (t transformer) treeTransform(s data.Iterable) data.Iterable {
	Print(data.Slice(s), "tree transform")
	if t.symbol == nil {
		return tree.MapToTree(tree.ObjectSlice(s), nil)
	}
	return tree.MapToTree(tree.ObjectSlice(s), nil, *t.symbol)
}

func treeMapping(n int, m processing.MappingFunction, opts TreeOutputOptions) processing.MappingFunction {
	return func(e interface{}) interface{} {
		o := e.(*tree.TreeObject)
		if o.Object != nil {
			return Fields(o.Graph+opts.ElemTitle(o), m(o.Object))
		}
		return Fields(o.Graph+" "+opts.NodeTitle(o), opts.NodeMapping(n, o)) // create empty table line
	}
}
