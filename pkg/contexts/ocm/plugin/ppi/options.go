package ppi

import (
	"encoding/json"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
)

type Options struct {
	Config json.RawMessage
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Config, "config", "c", nil, "plugin configuration")
}
