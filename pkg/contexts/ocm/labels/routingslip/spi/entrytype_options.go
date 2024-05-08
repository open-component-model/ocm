package spi

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets/flagsetscheme"
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
