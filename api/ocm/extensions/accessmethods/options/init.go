package options

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

var DefaultRegistry = flagsets.SetBaseTypes(flagsets.NewConfigOptionTypeRegistry())

func RegisterOption(o OptionType) OptionType {
	DefaultRegistry.RegisterOptionType(o)
	return o
}
