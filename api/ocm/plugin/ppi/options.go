package ppi

import (
	"encoding/json"

	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils/cobrautils/flag"
)

const (
	OptPluginConfig     = "config"
	OptPlugingLogConfig = "log-config"
)

type Options struct {
	Config    json.RawMessage
	LogConfig json.RawMessage
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Config, OptPluginConfig, "c", nil, "plugin configuration")
	flag.YAMLVarP(fs, &o.LogConfig, OptPlugingLogConfig, "", nil, "ocm logging configuration")
}
