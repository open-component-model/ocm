// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package inputs

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/modern-go/reflect2"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

////////////////////////////////////////////////////////////////////////////////

type Context interface {
	clictx.Context
	Printf(msg string, args ...interface{}) (int, error)
	Section(msg string, args ...interface{}) Context
	AddGap(gap string) Context
}

type context struct {
	clictx.Context
	printer common.Printer
}

func NewContext(ctx clictx.Context, pr common.Printer) Context {
	return &context{
		Context: ctx,
		printer: pr,
	}
}

func (c *context) Printf(msg string, args ...interface{}) (int, error) {
	return c.printer.Printf(msg, args...)
}

func (c *context) Section(msg string, args ...interface{}) Context {
	c.printer.Printf(msg+"\n", args...)
	return c.AddGap("  ")
}

func (c *context) AddGap(gap string) Context {
	return &context{
		Context: c.Context,
		printer: c.printer.AddGap(gap),
	}
}

type InputSpec interface {
	runtime.VersionedTypedObject
	Validate(fldPath *field.Path, ctx Context, inputFilePath string) field.ErrorList
	GetBlob(ctx Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error)
}

type InputType interface {
	runtime.TypedObjectDecoder
	runtime.VersionedTypedObject
	Usage() string
}

type DefaultInputType struct {
	runtime.ObjectVersionedType
	runtime.TypedObjectDecoder
	usage string
}

func NewInputType(name string, proto InputSpec, usage string) InputType {
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &DefaultInputType{
		ObjectVersionedType: runtime.NewVersionedObjectType(name),
		TypedObjectDecoder:  runtime.MustNewDirectDecoder(proto),
		usage:               usage,
	}
}

func (t *DefaultInputType) Usage() string {
	return t.usage
}

type InputTypeScheme interface {
	runtime.Scheme

	GetInputType(name string) InputType
	Register(name string, atype InputType)

	DecodeInputSpec(data []byte, unmarshaler runtime.Unmarshaler) (InputSpec, error)
	CreateInputSpec(obj runtime.TypedObject) (InputSpec, error)
}

type inputTypeScheme struct {
	runtime.SchemeBase
}

func NewInputTypeScheme(defaultRepoDecoder runtime.TypedObjectDecoder) InputTypeScheme {
	var rt InputSpec
	scheme := runtime.MustNewDefaultScheme(&rt, &UnknownInputSpec{}, false, defaultRepoDecoder)
	return &inputTypeScheme{scheme}
}

func (t *inputTypeScheme) AddKnownTypes(s InputTypeScheme) {
	t.SchemeBase.AddKnownTypes(s)
}

func (t *inputTypeScheme) GetInputType(name string) InputType {
	d := t.GetDecoder(name)
	if d == nil {
		return nil
	}
	return d.(InputType)
}

func (t *inputTypeScheme) RegisterByDecoder(name string, decoder runtime.TypedObjectDecoder) error {
	if _, ok := decoder.(InputType); !ok {
		return errors.ErrInvalid("type", reflect.TypeOf(decoder).String())
	}
	return t.SchemeBase.RegisterByDecoder(name, decoder)
}

func (t *inputTypeScheme) Register(name string, rtype InputType) {
	t.SchemeBase.RegisterByDecoder(name, rtype)
}

func (t *inputTypeScheme) DecodeInputSpec(data []byte, unmarshaler runtime.Unmarshaler) (InputSpec, error) {
	obj, err := t.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	if spec, ok := obj.(InputSpec); ok {
		return spec, nil
	}
	return nil, fmt.Errorf("invalid access spec type: yield %T instead of RepositorySpec", obj)
}

func (t *inputTypeScheme) CreateInputSpec(obj runtime.TypedObject) (InputSpec, error) {
	if s, ok := obj.(InputSpec); ok {
		r, err := t.SchemeBase.Convert(s)
		if err != nil {
			return nil, err
		}
		return r.(InputSpec), nil
	}
	if u, ok := obj.(*runtime.UnstructuredTypedObject); ok {
		raw, err := u.GetRaw()
		if err != nil {
			return nil, err
		}
		return t.DecodeInputSpec(raw, runtime.DefaultJSONEncoding)
	}
	return nil, fmt.Errorf("invalid object type %T for repository specs", obj)
}

// DefaultInputTypeScheme contains all globally known access serializer.
var DefaultInputTypeScheme = NewInputTypeScheme(nil)

func RegisterInputType(name string, atype InputType) {
	DefaultInputTypeScheme.Register(name, atype)
}

func CreateRepositorySpec(t runtime.TypedObject) (InputSpec, error) {
	return DefaultInputTypeScheme.CreateInputSpec(t)
}

////////////////////////////////////////////////////////////////////////////////

const ATTR_INPUT_TYPES = "github.com/open-component-model/ocm/cmds/ocm/common/inputs"

func For(ctx datacontext.Context) InputTypeScheme {
	if ctx == nil {
		return DefaultInputTypeScheme
	}
	return ctx.GetAttributes().GetAttribute(ATTR_INPUT_TYPES, DefaultInputTypeScheme).(InputTypeScheme)
}

func SetFor(ctx datacontext.Context, scheme InputTypeScheme) {
	ctx.GetAttributes().SetAttribute(ATTR_INPUT_TYPES, scheme)
}

////////////////////////////////////////////////////////////////////////////////

type UnknownInputSpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var _ InputSpec = &UnknownInputSpec{}

func (r *UnknownInputSpec) Validate(fldPath *field.Path, ctx Context, inputFilePath string) field.ErrorList {
	return field.ErrorList{field.Invalid(fldPath.Child("type"), r.GetType(), "unknown type")}
}

func (r *UnknownInputSpec) GetBlob(ctx Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	return nil, "", errors.ErrUnknown("input type", r.GetType())
}

////////////////////////////////////////////////////////////////////////////////

type GenericInputSpec struct {
	unstructured runtime.UnstructuredVersionedTypedObject
	effective    InputSpec
}

var _ InputSpec = &GenericInputSpec{}

func (s *GenericInputSpec) GetType() string {
	if s.effective != nil {
		return s.effective.GetType()
	}
	return s.unstructured.GetType()
}

func (s *GenericInputSpec) GetKind() string {
	if s.effective != nil {
		return s.effective.GetKind()
	}
	return s.unstructured.GetKind()
}

func (s *GenericInputSpec) GetVersion() string {
	if s.effective != nil {
		return s.effective.GetVersion()
	}
	return s.unstructured.GetVersion()
}

func (s *GenericInputSpec) Validate(fldPath *field.Path, ctx Context, inputFilePath string) field.ErrorList {
	if s.effective == nil {
		scheme := For(ctx)
		typeField := fldPath.Child("type")
		if s.GetType() == "" {
			return field.ErrorList{field.Required(typeField, "")}
		}
		if scheme.GetInputType(s.GetType()) == nil {
			return field.ErrorList{field.NotSupported(typeField, s.GetType(), scheme.KnownTypeNames())}
		}
		var err error
		s.effective, err = For(ctx).CreateInputSpec(s.unstructured)
		if err != nil {
			return field.ErrorList{field.InternalError(fldPath, err)}
		}
	}
	return s.effective.Validate(fldPath, ctx, inputFilePath)
}

func (s *GenericInputSpec) GetBlob(ctx Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	if s.effective == nil {
		var err error
		s.effective, err = s.Evaluate(For(ctx))
		if err != nil {
			return nil, "", err
		}
	}
	return s.effective.GetBlob(ctx, nv, inputFilePath)
}

func (s *GenericInputSpec) Evaluate(scheme InputTypeScheme) (InputSpec, error) {
	var err error
	if s == nil {
		return nil, nil
	}
	if s.effective == nil {
		var raw []byte
		raw, err = s.unstructured.GetRaw()
		if err != nil {
			return nil, err
		}
		s.effective, err = scheme.DecodeInputSpec(raw, runtime.DefaultJSONEncoding)
	}
	return s.effective, err
}

func (s GenericInputSpec) MarshalJSON() ([]byte, error) {
	if s.effective != nil {
		return json.Marshal(s.effective)
	}
	return s.unstructured.MarshalJSON()
}

func (s *GenericInputSpec) UnmarshalJSON(data []byte) error {
	s.effective = nil
	return s.unstructured.UnmarshalJSON(data)
}

func (s *GenericInputSpec) GetRaw() ([]byte, error) {
	if s.effective == nil {
		return json.Marshal(s.effective)
	}
	return s.unstructured.GetRaw()
}

func ToGenericInputSpec(spec InputSpec) (*GenericInputSpec, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	if g, ok := spec.(*GenericInputSpec); ok {
		return g, nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return newGenericInputSpec(data, runtime.DefaultJSONEncoding)
}

func NewGenericInputSpec(data []byte, unmarshaler runtime.Unmarshaler) (InputSpec, error) {
	s, err := newGenericInputSpec(data, unmarshaler)
	if err != nil {
		return nil, err // GO is great
	}
	return s, nil
}

func newGenericInputSpec(data []byte, unmarshaler runtime.Unmarshaler) (*GenericInputSpec, error) {
	gen := GenericInputSpec{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, &gen.unstructured)
	if err != nil {
		return nil, err
	}
	return &gen, nil
}

func Usage(scheme InputTypeScheme) string {
	s := `
The resource specification supports the following blob input types, specified
with the field <code>type</code> in the <code>input</code> field:
`
	for _, t := range scheme.KnownTypeNames() {
		s = fmt.Sprintf("%s\n- Input type <code>%s</code>\n\n%s", s, t, utils.IndentLines(scheme.GetInputType(t).Usage(), "  "))
	}
	return s + "\n"
}
