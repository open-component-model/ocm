package labelsel

import (
	"bytes"
	"container/list"
	"encoding/json"
	"reflect"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

func init() {
	logging.SetLevel(logging.ERROR, "yq-lib")
	yqlib.InitExpressionParser()
}

type Selector = selectors.LabelSelector

func Select(labels v1.Labels, sel ...Selector) (v1.Labels, error) {
	return selectors.SelectLabels(labels, sel...)
}

func Get(labels v1.Labels, sel ...Selector) v1.Labels {
	return selectors.GetLabels(labels, sel...)
}

////////////////////////////////////////////////////////////////////////////////

type name string

func (n name) MatchLabel(l *v1.Label) bool {
	return string(n) == l.Name
}

func Name(n string) *selectors.LabelSelectorImpl {
	return &selectors.LabelSelectorImpl{name(n)}
}

////////////////////////////////////////////////////////////////////////////////

type version string

func (n version) MatchLabel(l *v1.Label) bool {
	return string(n) == l.Version
}

func Version(n string) *selectors.LabelSelectorImpl {
	return &selectors.LabelSelectorImpl{version(n)}
}

///////////////////////////////////////////////////////////////////////////////

type signed bool

func (n signed) MatchLabel(l *v1.Label) bool {
	return bool(n) == l.Signing
}

func Signed(b ...bool) *selectors.LabelSelectorImpl {
	return &selectors.LabelSelectorImpl{signed(utils.OptionalDefaultedBool(true, b...))}
}

///////////////////////////////////////////////////////////////////////////////

type mergealgo string

func (n mergealgo) MatchLabel(l *v1.Label) bool {
	a := string(n)
	if l.Merge == nil {
		return a == ""
	}
	return a == l.Merge.Algorithm
}

func MergeAlgo(algo string) *selectors.LabelSelectorImpl {
	return &selectors.LabelSelectorImpl{mergealgo(algo)}
}

////////////////////////////////////////////////////////////////////////////////

func AsStructure(value interface{}) (interface{}, error) {
	var err error

	data, ok := value.([]byte)
	if !ok {
		data, err = json.Marshal(value)
		if err != nil {
			return nil, err
		}
	}

	var v interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Value matches a label by a label value.
// This selector should typically be combined with Name.
func Value(value interface{}) *selectors.LabelErrorSelectorImpl {
	data, err := AsStructure(value)
	return selectors.NewLabelErrorSelectorImpl(selectors.LabelSelectorFunc(func(l *v1.Label) bool {
		var value interface{}
		err := json.Unmarshal(l.Value, &value)
		if err != nil {
			return false
		}
		return reflect.DeepEqual(value, data)
	}), err)
}

////////////////////////////////////////////////////////////////////////////////

func YQParse(data []byte) (*yqlib.CandidateNode, error) {
	decoder := yqlib.NewYamlDecoder(yqlib.YamlPreferences{})
	err := decoder.Init(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return decoder.Decode()
}

type yqeval struct {
	expr  *yqlib.ExpressionNode
	value interface{}
}

func (v *yqeval) MatchLabel(l *v1.Label) bool {
	if v.expr == nil {
		return false
	}
	in, err := YQParse(l.Value)
	if err != nil {
		return false
	}
	t := yqlib.NewDataTreeNavigator()
	docs := list.New()
	docs.PushBack(in)
	context, err := t.GetMatchingNodes(yqlib.Context{MatchingNodes: docs}, v.expr)
	if err != nil {
		return false
	}
	if context.MatchingNodes.Len() != 1 {
		return false
	}
	data, err := context.MatchingNodes.Front().Value.(*yqlib.CandidateNode).MarshalJSON()
	if err != nil {
		return false
	}
	var out interface{}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(v.value, out)
}

// YQExpression matches a part of a label values described by a YQ expression.
// If value is a []byte, it is interpreted as JSON data, otherwise the value
// marshalled as JSON.
func YQExpression(expr string, value interface{}) *selectors.LabelErrorSelectorImpl {
	var data interface{}

	node, err := yqlib.ExpressionParser.ParseExpression(expr)
	if err == nil {
		data, err = AsStructure(value)
	}
	return selectors.NewLabelErrorSelectorImpl(&yqeval{node, data}, errors.Wrapf(err, "YQExpression selector"))
}
