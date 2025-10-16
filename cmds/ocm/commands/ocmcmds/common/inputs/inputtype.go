package inputs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/modern-go/reflect2"
	"k8s.io/apimachinery/pkg/util/validation/field"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

const KIND_INPUTTYPE = "input type"

////////////////////////////////////////////////////////////////////////////////

type Context interface {
	clictx.Context
	Printer() common.Printer
	Printf(msg string, args ...interface{}) (int, error)
	Variables() map[string]interface{}
	Section(msg string, args ...interface{}) Context
	AddGap(gap string) Context
}

type context struct {
	clictx.Context
	printer   common.Printer
	variables map[string]interface{}
}

func NewContext(ctx clictx.Context, pr common.Printer, variables map[string]interface{}) Context {
	return &context{
		Context:   ctx,
		printer:   pr,
		variables: variables,
	}
}

func (c *context) Printf(msg string, args ...interface{}) (int, error) {
	return c.printer.Printf(msg, args...)
}

func (c *context) Printer() common.Printer {
	return c.printer
}

func (c *context) Variables() map[string]interface{} {
	return c.variables
}

func (c *context) Section(msg string, args ...interface{}) Context {
	c.printer.Printf(msg+"\n", args...)
	return c.AddGap("  ")
}

func (c *context) AddGap(gap string) Context {
	return &context{
		Context:   c.Context,
		printer:   c.printer.AddGap(gap),
		variables: c.variables,
	}
}

type InputResourceInfo struct {
	// ComponentVersion is the name of the component version to generate.
	ComponentVersion common.NameVersion
	// ElementName is the name of the element to create.
	ElementName string
	// The path of the file the inputs description has been taken from.
	InputFilePath string
}

type InputSpec interface {
	runtime.VersionedTypedObject
	Validate(fldPath *field.Path, ctx Context, inputFilePath string) field.ErrorList
	GetBlob(ctx Context, info InputResourceInfo) (blobaccess.BlobAccess, string, error)
	GetInputVersion(ctx Context) string
}

type InputSpecBase struct {
	runtime.ObjectVersionedType `json:",inline"`
}

func (*InputSpecBase) GetInputVersion(ctx Context) string {
	return ""
}

type (
	InputSpecDecoder = runtime.TypedObjectDecoder[InputSpec]
)

type InputType interface {
	runtime.VersionedTypeInfo
	runtime.TypedObjectDecoder[InputSpec]

	ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler

	Usage() string
}

type DefaultInputType struct {
	runtime.ObjectVersionedType
	runtime.TypedObjectDecoder[InputSpec]
	usage      string
	clihandler flagsets.ConfigOptionTypeSetHandler
}

func NewInputType(name string, proto InputSpec, usage string, cfg flagsets.ConfigOptionTypeSetHandler) InputType {
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &DefaultInputType{
		ObjectVersionedType: runtime.NewVersionedTypedObject(name),
		TypedObjectDecoder:  runtime.MustNewDirectDecoder[InputSpec](proto),
		usage:               usage,
		clihandler:          cfg,
	}
}

func (t *DefaultInputType) ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler {
	return t.clihandler
}

func (t *DefaultInputType) Usage() string {
	group := ""
	if t.clihandler != nil {
		opts := t.clihandler.OptionTypeNames()
		var names []string
		if len(opts) > 0 {
			for _, o := range opts {
				names = append(names, "<code>--"+o+"</code>")
			}
			group = "\nOptions used to configure fields: " + strings.Join(names, ", ")
		}
	}
	return t.usage + group
}

func (t *DefaultInputType) ApplyConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if t.clihandler != nil {
		return t.clihandler.ApplyConfig(opts, config)
	}
	return nil
}

type InputTypeScheme interface {
	runtime.Scheme[InputSpec, InputType]

	ConfigTypeSetConfigProvider() flagsets.ConfigTypeOptionSetConfigProvider
	flagsets.ConfigProvider

	GetInputType(name string) InputType
	Register(atype InputType)

	GetInputSpecFor(opts flagsets.ConfigOptions) (InputSpec, error)
	DecodeInputSpec(data []byte, unmarshaler runtime.Unmarshaler) (InputSpec, error)
	CreateInputSpec(obj runtime.TypedObject) (InputSpec, error)
}

type inputTypeScheme struct {
	runtime.Scheme[InputSpec, InputType]
	optionTypes flagsets.ConfigTypeOptionSetConfigProvider
}

func NewInputTypeScheme(defaultRepoDecoder runtime.TypedObjectDecoder[InputSpec]) InputTypeScheme {
	scheme := runtime.MustNewDefaultScheme[InputSpec, InputType](&UnknownInputSpec{}, false, defaultRepoDecoder)
	prov := flagsets.NewTypedConfigProvider("input", "blob input specification", "inputType")
	prov.AddGroups("Input Specification Options")
	return &inputTypeScheme{scheme, prov}
}

func (t *inputTypeScheme) ConfigTypeSetConfigProvider() flagsets.ConfigTypeOptionSetConfigProvider {
	return t.optionTypes
}

func (t *inputTypeScheme) CreateOptions() flagsets.ConfigOptions {
	return t.optionTypes.CreateOptions()
}

func (t *inputTypeScheme) GetInputSpecFor(opts flagsets.ConfigOptions) (InputSpec, error) {
	cfg, err := t.GetConfigFor(opts)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return t.DecodeInputSpec(data, runtime.DefaultJSONEncoding)
}

func (t *inputTypeScheme) GetConfigFor(opts flagsets.ConfigOptions) (flagsets.Config, error) {
	return t.optionTypes.GetConfigFor(opts)
}

func (t *inputTypeScheme) GetInputType(name string) InputType {
	d := t.GetDecoder(name)
	if d == nil {
		return nil
	}
	return d
}

func (t *inputTypeScheme) Register(rtype InputType) {
	if rtype == nil {
		return
	}
	t.RegisterByDecoder(rtype.GetType(), rtype)
	t.optionTypes.AddTypeSet(rtype.ConfigOptionTypeSetHandler())
}

func (t *inputTypeScheme) DecodeInputSpec(data []byte, unmarshaler runtime.Unmarshaler) (InputSpec, error) {
	return t.Decode(data, unmarshaler)
}

func (t *inputTypeScheme) CreateInputSpec(obj runtime.TypedObject) (InputSpec, error) {
	if s, ok := obj.(InputSpec); ok {
		r, err := t.Convert(s)
		if err != nil {
			return nil, err
		}
		return r, nil
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

func RegisterInputType(atype InputType) {
	DefaultInputTypeScheme.Register(atype)
}

func CreateRepositorySpec(t runtime.TypedObject) (InputSpec, error) {
	return DefaultInputTypeScheme.CreateInputSpec(t)
}

////////////////////////////////////////////////////////////////////////////////

const ATTR_INPUT_TYPES = "ocm.software/ocm/cmds/ocm/common/inputs"

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

func (r *UnknownInputSpec) GetBlob(ctx Context, info InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	return nil, "", errors.ErrUnknown("input type", r.GetType())
}

func (s *UnknownInputSpec) GetInputVersion(ctx Context) string {
	return ""
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

func (s *GenericInputSpec) GetBlob(ctx Context, info InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	if s.effective == nil {
		var err error
		s.effective, err = s.Evaluate(For(ctx))
		if err != nil {
			return nil, "", err
		}
	}
	return s.effective.GetBlob(ctx, info)
}

func (s *GenericInputSpec) GetInputVersion(ctx Context) string {
	if s.effective == nil {
		var err error
		s.effective, err = s.Evaluate(For(ctx))
		if err != nil {
			return ""
		}
	}
	return s.effective.GetInputVersion(ctx)
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
with the field <code>type</code> in the <code>input</code> field:`
	for _, t := range scheme.KnownTypeNames() {
		s = fmt.Sprintf("%s\n\n- Input type <code>%s</code>\n\n%s", s, t, utils.IndentLines(scheme.GetInputType(t).Usage(), "  "))
	}
	return s + "\n"
}
