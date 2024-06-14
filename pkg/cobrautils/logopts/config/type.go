package config

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/open-component-model/ocm/pkg/cobrautils/logopts"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "logfile.ocm" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based config interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`

	LogFileName string `json:"logFileName"`
}

// New creates a new memory ConfigSpec.
func New(logfile string) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		LogFileName:         logfile,
	}
}

func (c *Config) GetType() string {
	return ConfigType
}

func (c *Config) ApplyTo(_ config.Context, target interface{}) error {
	if logopts.GlobalLogFileOverride == c.LogFileName {
		return nil
	}
	if c.LogFileName != "" {
		logfile, err := logopts.LogFileFor(c.LogFileName, osfs.OsFs)
		if err != nil {
			return errors.Wrapf(err, "cannot open log file %q", c.LogFileName)
		}
		logopts.ConfigureLogrusFor(logging.DefaultContext(), false, logfile)
		logopts.GlobalLogFile = logfile
		logopts.GlobalLogFileOverride = c.LogFileName
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> is used to override
the logging destination and enforce the logging to a dedicated log file.

<pre>
    type: ` + ConfigType + `
    logFile: /tmp/ocm-cli-log-0815
</pre>
`
