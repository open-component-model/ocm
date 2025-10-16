package logopts

import (
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/config"
	"github.com/mandelsoft/logging/logrusl/adapter"
	"github.com/mandelsoft/logging/logrusr"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/datacontext/attrs/logforward"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils"
	logdata "ocm.software/ocm/api/utils/cobrautils/logopts/logging"
)

// ConfigFragment is a serializable log config used
// for CLI commands.
type ConfigFragment struct {
	LogLevel  string   `json:"logLevel,omitempty"`
	LogConfig string   `json:"logConfig,omitempty"`
	LogKeys   []string `json:"logKeys,omitempty"`
	Json      bool     `json:"json,omitempty"`

	// LogFileName is a CLI option, only. Do not serialize and forward
	LogFileName string `json:"-"`
}

func (c *ConfigFragment) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&c.Json, "logJson", "", false, "log as json instead of human readable logs")
	fs.StringVarP(&c.LogLevel, "loglevel", "l", "", "set log level")
	fs.StringVarP(&c.LogFileName, "logfile", "L", "", "set log file")
	fs.StringVarP(&c.LogConfig, "logconfig", "", "", "log config")
	fs.StringArrayVarP(&c.LogKeys, "logkeys", "", nil, "log tags/realms(with leading /) to be enabled ([/[+]]name{,[/[+]]name}[=level])")
}

func (c *ConfigFragment) GetLogConfig(fss ...vfs.FileSystem) (*config.Config, error) {
	var (
		err error
		cfg *config.Config
	)

	if c.LogConfig != "" {
		var data []byte
		if strings.HasPrefix(c.LogConfig, "@") {
			data, err = utils.ReadFile(c.LogConfig[1:], utils.FileSystem(fss...))
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read logging config file %q", c.LogConfig[1:])
			}
		} else {
			data = []byte(c.LogConfig)
		}
		if cfg, err = config.EvaluateFromData(data); err != nil {
			return nil, errors.Wrapf(err, "invalid logging config: %q", c.LogConfig)
		}
	} else {
		cfg = &config.Config{DefaultLevel: "Warn"}
	}

	for _, t := range c.LogKeys {
		level := logging.InfoLevel
		i := strings.Index(t, "=")
		if i >= 0 {
			level, err = logging.ParseLevel(t[i+1:])
			if err != nil {
				return nil, errors.Wrapf(err, "invalid log tag setting")
			}
			t = t[:i]
		}
		var cfgcond []config.Condition

		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if strings.HasPrefix(tag, "/") {
				realm := tag[1:]
				if strings.HasPrefix(realm, "+") {
					cfgcond = append(cfgcond, config.RealmPrefix(realm[1:]))
				} else {
					cfgcond = append(cfgcond, config.Realm(realm))
				}
			} else {
				cfgcond = append(cfgcond, config.Tag(tag))
			}
		}
		cfg.Rules = append(cfg.Rules, config.ConditionalRule(logging.LevelName(level), cfgcond...))
	}

	if c.LogLevel != "" {
		_, err := logging.ParseLevel(c.LogLevel)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid log level %q", c.LogLevel)
		}
		cfg.DefaultLevel = c.LogLevel
	}

	return cfg, nil
}

func (c *ConfigFragment) Evaluate(ctx ocm.Context, logctx logging.Context, main bool) (*EvaluatedOptions, error) {
	var err error
	var opts EvaluatedOptions

	for logctx.Tree().GetBaseContext() != nil {
		logctx = logctx.Tree().GetBaseContext()
	}

	fs := vfsattr.Get(ctx)
	if main && c.LogFileName != "" && logdata.GlobalLogFileOverride == "" {
		if opts.LogFile == nil {
			opts.LogFile, err = logdata.LogFileFor(c.LogFileName, fs)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot open log file %q", opts.LogFile)
			}
		}
		logdata.ConfigureLogrusFor(logctx, !c.Json, opts.LogFile)
		if logctx == logging.DefaultContext() {
			logdata.GlobalLogFile = opts.LogFile
		}
	} else {
		// overwrite current log formatter in case of a logrus logger is
		// used as logging backend.
		var f logrus.Formatter = adapter.NewJSONFormatter()
		if !c.Json {
			f = adapter.NewTextFmtFormatter()
		}
		logrusr.SetFormatter(logging.UnwrapLogSink(logctx.GetSink()), f)
	}

	cfg, err := c.GetLogConfig(fs)
	if err != nil {
		return &opts, err
	}
	err = config.Configure(logctx, cfg)
	if err != nil {
		return &opts, err
	}
	opts.LogForward = cfg
	logforward.Set(ctx.AttributesContext(), opts.LogForward)

	return &opts, nil
}
