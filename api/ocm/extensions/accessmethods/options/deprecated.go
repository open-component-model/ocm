package options

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

const (
	// Deprecated: use package [flagsets].
	TYPE_STRING = flagsets.TYPE_STRING
	// Deprecated: use package [flagsets].
	TYPE_STRINGARRAY = flagsets.TYPE_STRINGARRAY
	// Deprecated: use package [flagsets].
	TYPE_STRING2STRING = flagsets.TYPE_STRING2STRING
	// Deprecated: use package [flagsets].
	TYPE_INT = flagsets.TYPE_INT
	// Deprecated: use package [flagsets].
	TYPE_BOOL = flagsets.TYPE_BOOL
	// Deprecated: use package [flagsets].
	TYPE_YAML = flagsets.TYPE_YAML
	// Deprecated: use package [flagsets].
	TYPE_STRINGMAPYAML = flagsets.TYPE_STRINGMAPYAML
	// Deprecated: use package [flagsets].
	TYPE_STRING2YAML = flagsets.TYPE_STRING2YAML
	// Deprecated: use package [flagsets].
	TYPE_STRING2STRINGSLICE = flagsets.TYPE_STRING2STRINGSLICE
	// Deprecated: use package [flagsets].
	TYPE_STRINGCOLONSTRINGSLICE = flagsets.TYPE_STRINGCOLONSTRINGSLICE
	// Deprecated: use package [flagsets].
	TYPE_BYTES = flagsets.TYPE_BYTES
	// Deprecated: use package [flagsets].
	TYPE_IDENTITYPATH = flagsets.TYPE_IDENTITYPATH
)

// Deprecated: use packagge [flagsets].
type OptionType = flagsets.ConfigOptionType

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use packagge [flagsets].
func NewStringOptionType(name, desc string) OptionType {
	return flagsets.NewStringOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewStringArrayOptionType(name, desc string) OptionType {
	return flagsets.NewStringArrayOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewIntOptionType(name, desc string) OptionType {
	return flagsets.NewIntOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewBoolOptionType(name, desc string) OptionType {
	return flagsets.NewBoolOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewYAMLOptionType(name, desc string) OptionType {
	return flagsets.NewYAMLOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewValueMapYAMLOptionType(name, desc string) OptionType {
	return flagsets.NewValueMapYAMLOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewValueMapOptionType(name, desc string) OptionType {
	return flagsets.NewValueMapOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewStringMapOptionType(name, desc string) OptionType {
	return flagsets.NewStringMapOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewStringSliceMapOptionType(name, desc string) OptionType {
	return flagsets.NewStringSliceMapOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewStringSliceMapColonOptionType(name, desc string) OptionType {
	return flagsets.NewStringSliceMapColonOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewBytesOptionType(name, desc string) OptionType {
	return flagsets.NewBytesOptionType(name, desc)
}

// Deprecated: use packagge [flagsets].
func NewIdentityPathOptionType(name, desc string) OptionType {
	return flagsets.NewIdentityPathOptionType(name, desc)
}
