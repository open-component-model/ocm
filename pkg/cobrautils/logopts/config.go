package logopts

import (
	"runtime"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/config"
	"github.com/mandelsoft/logging/logrusr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/logforward"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/utils"
)

// ConfigFragment is a serializable log config used
// for CLI commands.
type ConfigFragment struct {
	LogLevel    string   `json:"logLevel,omitempty"`
	LogFileName string   `json:"logFileName,omitempty"`
	LogConfig   string   `json:"logConfig,omitempty"`
	LogKeys     []string `json:"logKeys,omitempty"`
}

func (c *ConfigFragment) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&c.LogLevel, "loglevel", "l", "", "set log level")
	fs.StringVarP(&c.LogFileName, "logfile", "L", "", "set log file")
	fs.StringVarP(&c.LogConfig, "logconfig", "", "", "log config")
	fs.StringArrayVarP(&c.LogKeys, "logkeys", "", nil, "log tags/realms(with leading /) to be enabled ([/[+]]name{,[/[+]]name}[=level])")
}

func (c *ConfigFragment) Evaluate(ctx ocm.Context, logctx logging.Context) (*EvaluatedOptions, error) {
	var err error
	var opts EvaluatedOptions

	for logctx.Tree().GetBaseContext() != nil {
		logctx = logctx.Tree().GetBaseContext()
	}

	if c.LogLevel != "" {
		l, err := logging.ParseLevel(c.LogLevel)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid log level %q", c.LogLevel)
		}
		logctx.SetDefaultLevel(l)
	} else {
		logctx.SetDefaultLevel(logging.WarnLevel)
	}
	opts.LogForward = &config.Config{DefaultLevel: logging.LevelName(logctx.GetDefaultLevel())}

	fs := vfsattr.Get(ctx)
	if c.LogFileName != "" {
		if opts.LogFile == nil {
			opts.LogFile, err = LogFileFor(c.LogFileName, fs)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot open log file %q", opts.LogFile)
			}
		}
		log := logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
		log.SetOutput(opts.LogFile.File())
		logctx.SetBaseLogger(logrusr.New(log))
		runtime.SetFinalizer(log, func(_ *logrus.Logger) { opts.LogFile.Close() })
	}

	if c.LogConfig != "" {
		var cfg []byte
		if strings.HasPrefix(c.LogConfig, "@") {
			cfg, err = utils.ReadFile(c.LogConfig[1:], fs)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read logging config file %q", c.LogConfig[1:])
			}
		} else {
			cfg = []byte(c.LogConfig)
		}
		if err = config.ConfigureWithData(logctx, cfg); err != nil {
			return nil, errors.Wrapf(err, "invalid logging config: %q", c.LogConfig)
		}
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
		var cond []logging.Condition
		var cfgcond []config.Condition

		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if strings.HasPrefix(tag, "/") {
				realm := tag[1:]
				if strings.HasPrefix(realm, "+") {
					cond = append(cond, logging.NewRealmPrefix(realm[1:]))
					cfgcond = append(cfgcond, config.RealmPrefix(realm[1:]))
				} else {
					cond = append(cond, logging.NewRealm(realm))
					cfgcond = append(cfgcond, config.Realm(realm))
				}
			} else {
				cond = append(cond, logging.NewTag(tag))
				cfgcond = append(cfgcond, config.Tag(tag))
			}
		}
		rule := logging.NewConditionRule(level, cond...)
		opts.LogForward.Rules = append(opts.LogForward.Rules, config.ConditionalRule(logging.LevelName(level), cfgcond...))
		logctx.AddRule(rule)
	}
	logforward.Set(ctx.AttributesContext(), opts.LogForward)

	return &opts, nil
}
