package spi

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/cobrautils/flagsets/flagsetscheme"
)

type EntryTypeOption = flagsetscheme.TypeOption

func WithFormatSpec(value string) EntryTypeOption {
	return flagsetscheme.WithFormatSpec(value)
}

func WithDescription(value string) EntryTypeOption {
	return flagsetscheme.WithDescription(value)
}

func WithConfigHandler(value flagsets.ConfigOptionTypeSetHandler) EntryTypeOption {
	return flagsetscheme.WithConfigHandler(value)
}
