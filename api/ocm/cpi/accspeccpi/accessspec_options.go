package accspeccpi

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/cobrautils/flagsets/flagsetscheme"
)

type AccessSpecTypeOption = flagsetscheme.TypeOption

func WithFormatSpec(value string) AccessSpecTypeOption {
	return flagsetscheme.WithFormatSpec(value)
}

func WithDescription(value string) AccessSpecTypeOption {
	return flagsetscheme.WithDescription(value)
}

func WithConfigHandler(value flagsets.ConfigOptionTypeSetHandler) AccessSpecTypeOption {
	return flagsetscheme.WithConfigHandler(value)
}
