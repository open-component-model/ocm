package flagsets

import (
	"fmt"

	"github.com/spf13/pflag"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/api/utils/cobrautils/groups"
)

const (
	TYPE_STRING                 = "string"
	TYPE_STRINGARRAY            = "[]string"
	TYPE_STRING2STRING          = "string=string"
	TYPE_INT                    = "int"
	TYPE_BOOL                   = "bool"
	TYPE_YAML                   = "YAML"
	TYPE_STRINGMAPYAML          = "map[string]YAML"
	TYPE_STRING2YAML            = "string=YAML"
	TYPE_STRING2STRINGSLICE     = "string=string,string"
	TYPE_STRINGCOLONSTRINGSLICE = "string:string,string"
	TYPE_BYTES                  = "[]byte"
	TYPE_IDENTITYPATH           = "[]identity"
)

func SetBaseTypes(r ConfigOptionTypeRegistry) ConfigOptionTypeRegistry {
	r.RegisterValueType(TYPE_STRING, NewStringOptionType, "string value")
	r.RegisterValueType(TYPE_STRINGARRAY, NewStringArrayOptionType, "list of string values")
	r.RegisterValueType(TYPE_STRING2STRING, NewStringMapOptionType, "string map defined by dedicated assignments")
	r.RegisterValueType(TYPE_INT, NewIntOptionType, "integer value")
	r.RegisterValueType(TYPE_BOOL, NewBoolOptionType, "boolean flag")
	r.RegisterValueType(TYPE_YAML, NewYAMLOptionType, "JSON or YAML document string")
	r.RegisterValueType(TYPE_STRINGMAPYAML, NewValueMapYAMLOptionType, "JSON or YAML map")
	r.RegisterValueType(TYPE_STRING2YAML, NewValueMapOptionType, "string map with arbitrary values defined by dedicated assignments")
	r.RegisterValueType(TYPE_STRING2STRINGSLICE, NewStringSliceMapOptionType, "string map defined by dedicated assignment of comma separated strings")
	r.RegisterValueType(TYPE_STRINGCOLONSTRINGSLICE, NewStringSliceMapColonOptionType, "string map defined by dedicated assignment of comma separated strings")
	r.RegisterValueType(TYPE_BYTES, NewBytesOptionType, "byte value")
	r.RegisterValueType(TYPE_IDENTITYPATH, NewIdentityPathOptionType, "identity path")
	return r
}

type TypeOptionBase struct {
	name        string
	description string
	valueType   string
}

func (b *TypeOptionBase) GetName() string {
	return b.name
}

func (b *TypeOptionBase) GetDescription() string {
	return fmt.Sprintf("[*%s*] %s", b.ValueType(), b.description)
}

func (b *TypeOptionBase) GetDescriptionText() string {
	return b.description
}

func (o *TypeOptionBase) ValueType() string {
	return o.valueType
}

func (o *TypeOptionBase) Equal(t ConfigOptionType) bool {
	return o.name == t.GetName() &&
		o.description == t.GetDescriptionText() &&
		o.valueType == t.ValueType()
}

////////////////////////////////////////////////////////////////////////////////

type OptionBase struct {
	otyp   ConfigOptionType
	flag   *pflag.Flag
	groups []string
}

func NewOptionBase(otyp ConfigOptionType) OptionBase {
	return OptionBase{otyp: otyp}
}

func (b *OptionBase) Type() ConfigOptionType {
	return b.otyp
}

func (b *OptionBase) GetName() string {
	return b.otyp.GetName()
}

func (b *OptionBase) Description() string {
	return b.otyp.GetDescriptionText()
}

func (b *OptionBase) Changed() bool {
	return b.flag.Changed
}

func (b *OptionBase) AddGroups(groups ...string) {
	b.groups = AddGroups(b.groups, groups...)
	b.addGroups()
}

func (b *OptionBase) addGroups() {
	if len(b.groups) == 0 || b.flag == nil {
		return
	}
	if b.flag.Annotations == nil {
		b.flag.Annotations = map[string][]string{}
	}
	list := b.flag.Annotations[groups.FlagGroupAnnotation]
	b.flag.Annotations[groups.FlagGroupAnnotation] = AddGroups(list, b.groups...)
}

func (b *OptionBase) TweakFlag(f *pflag.Flag) {
	b.flag = f
	b.addGroups()
}

////////////////////////////////////////////////////////////////////////////////

type StringOptionType struct {
	TypeOptionBase
}

func NewStringOptionType(name string, description string) ConfigOptionType {
	return &StringOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_STRING},
	}
}

func (s *StringOptionType) Create() Option {
	return &StringOption{
		OptionBase: NewOptionBase(s),
	}
}

type StringOption struct {
	OptionBase
	value string
}

var _ Option = (*StringOption)(nil)

func (o *StringOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.StringVarPF(fs, &o.value, o.otyp.GetName(), "", "", o.otyp.GetDescriptionText()))
}

func (o *StringOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StringArrayOptionType struct {
	TypeOptionBase
}

func NewStringArrayOptionType(name string, description string) ConfigOptionType {
	return &StringArrayOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_STRINGARRAY},
	}
}

func (s *StringArrayOptionType) Create() Option {
	return &StringArrayOption{
		OptionBase: NewOptionBase(s),
	}
}

type StringArrayOption struct {
	OptionBase
	value []string
}

var _ Option = (*StringArrayOption)(nil)

func (o *StringArrayOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.StringArrayVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *StringArrayOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////

type BoolOptionType struct {
	TypeOptionBase
}

func NewBoolOptionType(name string, description string) ConfigOptionType {
	return &BoolOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_BOOL},
	}
}

func (s *BoolOptionType) Create() Option {
	return &BoolOption{
		OptionBase: NewOptionBase(s),
	}
}

type BoolOption struct {
	OptionBase
	value bool
}

var _ Option = (*BoolOption)(nil)

func (o *BoolOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.BoolVarPF(fs, &o.value, o.otyp.GetName(), "", false, o.otyp.GetDescriptionText()))
}

func (o *BoolOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type IntOptionType struct {
	TypeOptionBase
}

func NewIntOptionType(name string, description string) ConfigOptionType {
	return &IntOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_INT},
	}
}

func (s *IntOptionType) Create() Option {
	return &IntOption{
		OptionBase: NewOptionBase(s),
	}
}

type IntOption struct {
	OptionBase
	value int
}

var _ Option = (*IntOption)(nil)

func (o *IntOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.IntVarPF(fs, &o.value, o.otyp.GetName(), "", 0, o.otyp.GetDescriptionText()))
}

func (o *IntOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type YAMLOptionType struct {
	TypeOptionBase
}

func NewYAMLOptionType(name string, description string) ConfigOptionType {
	return &YAMLOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_YAML},
	}
}

func (s *YAMLOptionType) Create() Option {
	return &YAMLOption{
		OptionBase: NewOptionBase(s),
	}
}

type YAMLOption struct {
	OptionBase
	value interface{}
}

var _ Option = (*YAMLOption)(nil)

func (o *YAMLOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.YAMLVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *YAMLOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ValueMapYAMLOptionType struct {
	TypeOptionBase
}

func NewValueMapYAMLOptionType(name string, description string) ConfigOptionType {
	return &ValueMapYAMLOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_STRINGMAPYAML},
	}
}

func (s *ValueMapYAMLOptionType) Create() Option {
	return &ValueMapYAMLOption{
		OptionBase: NewOptionBase(s),
	}
}

type ValueMapYAMLOption struct {
	OptionBase
	value map[string]interface{}
}

var _ Option = (*ValueMapYAMLOption)(nil)

func (o *ValueMapYAMLOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.YAMLVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *ValueMapYAMLOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ValueMapOptionType struct {
	TypeOptionBase
}

func NewValueMapOptionType(name string, description string) ConfigOptionType {
	return &ValueMapOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_STRING2YAML},
	}
}

func (s *ValueMapOptionType) Create() Option {
	return &ValueMapOption{
		OptionBase: NewOptionBase(s),
	}
}

type ValueMapOption struct {
	OptionBase
	value map[string]interface{}
}

var _ Option = (*ValueMapOption)(nil)

func (o *ValueMapOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.StringToValueVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *ValueMapOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StringMapOptionType struct {
	TypeOptionBase
}

func NewStringMapOptionType(name string, description string) ConfigOptionType {
	return &StringMapOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_STRING2STRING},
	}
}

func (s *StringMapOptionType) Create() Option {
	return &StringMapOption{
		OptionBase: NewOptionBase(s),
	}
}

type StringMapOption struct {
	OptionBase
	value map[string]string
}

var _ Option = (*StringMapOption)(nil)

func (o *StringMapOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.StringToStringVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *StringMapOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type BytesOptionType struct {
	TypeOptionBase
}

func NewBytesOptionType(name string, description string) ConfigOptionType {
	return &BytesOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_BYTES},
	}
}

func (s *BytesOptionType) Create() Option {
	return &BytesOption{
		OptionBase: NewOptionBase(s),
	}
}

type BytesOption struct {
	OptionBase
	value []byte
}

var _ Option = (*BytesOption)(nil)

func (o *BytesOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.BytesBase64VarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *BytesOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StringSliceMapOptionType struct {
	TypeOptionBase
}

func NewStringSliceMapOptionType(name string, description string) ConfigOptionType {
	return &StringSliceMapOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_STRING2STRINGSLICE},
	}
}

func (s *StringSliceMapOptionType) Create() Option {
	return &StringSliceMapOption{
		OptionBase: NewOptionBase(s),
	}
}

type StringSliceMapOption struct {
	OptionBase
	value map[string][]string
}

var _ Option = (*StringSliceMapOption)(nil)

func (o *StringSliceMapOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.StringToStringSliceVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *StringSliceMapOption) Value() interface{} {
	return o.value
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StringSliceMapColonOptionType struct {
	TypeOptionBase
}

func NewStringSliceMapColonOptionType(name string, description string) ConfigOptionType {
	return &StringSliceMapColonOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_STRINGCOLONSTRINGSLICE},
	}
}

func (s *StringSliceMapColonOptionType) Create() Option {
	return &StringSliceMapColonOption{
		OptionBase: NewOptionBase(s),
	}
}

type StringSliceMapColonOption struct {
	OptionBase
	value map[string][]string
}

var _ Option = (*StringSliceMapColonOption)(nil)

func (o *StringSliceMapColonOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.StringColonStringSliceVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *StringSliceMapColonOption) Value() interface{} {
	return o.value
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type IdentityPathOptionType struct {
	TypeOptionBase
}

func NewIdentityPathOptionType(name string, description string) ConfigOptionType {
	return &IdentityPathOptionType{
		TypeOptionBase: TypeOptionBase{name, description, TYPE_IDENTITYPATH},
	}
}

func (s *IdentityPathOptionType) Create() Option {
	return &IdentityPathOption{
		OptionBase: NewOptionBase(s),
	}
}

type IdentityPathOption struct {
	OptionBase
	value []map[string]string
}

var _ Option = (*IdentityPathOption)(nil)

func (o *IdentityPathOption) AddFlags(fs *pflag.FlagSet) {
	o.TweakFlag(flag.IdentityPathVarPF(fs, &o.value, o.otyp.GetName(), "", nil, o.otyp.GetDescriptionText()))
}

func (o *IdentityPathOption) Value() interface{} {
	var result []v1.Identity
	for _, v := range o.value {
		result = append(result, v1.Identity(v))
	}
	return result
}
