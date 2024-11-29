package options

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

var DefaultRegistry = flagsets.SetBaseTypes(flagsets.NewConfigOptionTypeRegistry())

func RegisterOption(o flagsets.ConfigOptionType) flagsets.ConfigOptionType {
	DefaultRegistry.RegisterOptionType(o)
	return o
}
