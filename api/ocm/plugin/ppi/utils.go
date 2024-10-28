package ppi

import (
	"encoding/json"
	"reflect"

	errors2 "github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/utils/runtime"
)

type decoder runtime.TypedObjectDecoder[runtime.TypedObject]

const KIND_QUESTION = "question"

type AccessMethodBase struct {
	decoder
	nameDescription

	version string
	format  string
}

func MustNewAccessMethodBase(name, version string, proto AccessSpec, desc string, format string) AccessMethodBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}

	return AccessMethodBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		version: version,
		format:  format,
	}
}

func (b *AccessMethodBase) Version() string {
	return b.version
}

func (b *AccessMethodBase) Format() string {
	return b.format
}

////////////////////////////////////////////////////////////////////////////////

type UploaderBase = nameDescription

func MustNewUploaderBase(name, desc string) UploaderBase {
	return UploaderBase{
		name: name,
		desc: desc,
	}
}

////////////////////////////////////////////////////////////////////////////////

type ValueSetBase struct {
	decoder
	nameDescription

	version string
	format  string

	purposes []string
}

func MustNewValueSetBase(name, version string, proto runtime.TypedObject, purposes []string, desc string, format string) ValueSetBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}
	return ValueSetBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		version:  version,
		format:   format,
		purposes: slices.Clone(purposes),
	}
}

func (b *ValueSetBase) Version() string {
	return b.version
}

func (b *ValueSetBase) Format() string {
	return b.format
}

func (b *ValueSetBase) Purposes() []string {
	return b.purposes
}

////////////////////////////////////////////////////////////////////////////////

type nameDescription struct {
	name string
	desc string
}

func (b *nameDescription) Name() string {
	return b.name
}

func (b *nameDescription) Description() string {
	return b.desc
}

////////////////////////////////////////////////////////////////////////////////

type transferHandler struct {
	name        string
	description string
	questions   []DecisionHandler
}

func NewTransferHandler(name, desc string) *transferHandler {
	return &transferHandler{
		name:        name,
		description: desc,
		questions:   nil,
	}
}

func (t *transferHandler) GetName() string {
	return t.name
}

func (t *transferHandler) GetDescription() string {
	return t.description
}

func (t *transferHandler) GetQuestions() []DecisionHandler {
	return t.questions
}

func (t *transferHandler) RegisterDecision(h DecisionHandler) error {
	if TransferHandlerQuestions[h.GetQuestion()] == nil {
		return errors2.ErrInvalid(KIND_QUESTION, h.GetQuestion())
	}
	for _, e := range t.questions {
		if e.GetQuestion() == h.GetQuestion() {
			return errors2.ErrAlreadyExists(KIND_QUESTION, e.GetQuestion())
		}
	}
	t.questions = append(t.questions, h)
	return nil
}

// DecisionHandlerBase provides access to the
// non-functional attributes of a DecisionHandler.
// It can be created with NewDecisionHandlerBase and
// embedded into the final DecisionHandler implementation.
type DecisionHandlerBase struct {
	question    string
	description string
	labels      *[]string
}

func (d *DecisionHandlerBase) GetQuestion() string {
	return d.question
}

func (d *DecisionHandlerBase) GetDescription() string {
	return d.description
}

func (d *DecisionHandlerBase) GetLabels() *[]string {
	return d.labels
}

func NewDecisionHandlerBase(q, desc string, labels ...string) DecisionHandlerBase {
	return DecisionHandlerBase{q, desc, generics.Pointer(slices.Clone(labels))}
}

////////////////////////////////////////////////////////////////////////////////

// Config is a generic structured config stored in a string map.
type Config map[string]interface{}

func (c Config) GetValue(name string) (interface{}, bool) {
	v, ok := c[name]
	return v, ok
}

func (c Config) ConvertFor(list ...options.OptionType) error {
	for _, o := range list {
		if v, ok := c[o.GetName()]; ok {
			t := reflect.TypeOf(o.Create().Value())
			if t != reflect.TypeOf(v) {
				data, err := json.Marshal(v)
				if err != nil {
					return errors.Wrapf(err, "cannot marshal option value for %q", o.GetName())
				}
				value := reflect.New(t)
				err = json.Unmarshal(data, value.Interface())
				if err != nil {
					return errors.Wrapf(err, "cannot unmarshal option value for %q[%s]", o.GetName(), o.ValueType())
				}
				c[o.GetName()] = value.Elem().Interface()
			}
		}
	}
	return nil
}
