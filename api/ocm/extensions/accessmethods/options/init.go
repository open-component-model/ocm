package options

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
)

func init() {
	DefaultRegistry.RegisterValueType(TYPE_STRING, NewStringOptionType, "string value")
	DefaultRegistry.RegisterValueType(TYPE_STRINGARRAY, NewStringArrayOptionType, "list of string values")
	DefaultRegistry.RegisterValueType(TYPE_STRING2STRING, NewStringMapOptionType, "string map defined by dedicated assignments")
	DefaultRegistry.RegisterValueType(TYPE_INT, NewIntOptionType, "integer value")
	DefaultRegistry.RegisterValueType(TYPE_BOOL, NewBoolOptionType, "boolean flag")
	DefaultRegistry.RegisterValueType(TYPE_YAML, NewYAMLOptionType, "JSON or YAML document string")
	DefaultRegistry.RegisterValueType(TYPE_STRINGMAPYAML, NewValueMapYAMLOptionType, "JSON or YAML map")
	DefaultRegistry.RegisterValueType(TYPE_STRING2YAML, NewValueMapOptionType, "string map with arbitrary values defined by dedicated assignments")
	DefaultRegistry.RegisterValueType(TYPE_STRING2STRINGSLICE, NewStringSliceMapOptionType, "string map defined by dedicated assignment of comma separated strings")
	DefaultRegistry.RegisterValueType(TYPE_STRINGCOLONSTRINGSLICE, NewStringSliceMapColonOptionType, "string map defined by dedicated assignment of comma separated strings")
	DefaultRegistry.RegisterValueType(TYPE_BYTES, NewBytesOptionType, "byte value")
}

func RegisterOption(o OptionType) OptionType {
	DefaultRegistry.RegisterOptionType(o)
	return o
}
