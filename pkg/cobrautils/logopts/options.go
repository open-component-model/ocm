// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package logopts

import (
	"encoding/json"
	"strings"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/config"
	"github.com/mandelsoft/logging/logrusr"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/logforward"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
)

var Description = `
The <code>--log*</code> options can be used to configure the logging behaviour.
There is a quick config option <code>--log-keys</code> to configure simple
tag/realm based condition rules. The comma-separated names build an AND rule.
Hereby, names starting with a slash (<code>/</code>) denote a realm (without the leading slash).
A realm is a slash separated sequence of identifiers, which matches all logging realms
with the given realms as path prefix. A tag directly matches the logging tags.
Used tags and realms can be found under topic <CMD>ocm logging</CMD>. The ocm coding basically
uses the realm <code>ocm</code>.
The default level to enable is <code>info</code>. Separated by an equal sign (<code>=</code>)
optiobally a dedicated level can be specified. Log levels can be (<code>error</code>,
<code>warn</code>, <code>info</code>, <code>debug</code> and <code>trace</code>.
The default level is <code>warn</code>.
`

type Options struct {
	LogLevel    string
	LogFileName string
	LogConfig   string
	LogKeys     []string

	LogFile    vfs.File
	LogForward *config.Config
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.LogLevel, "loglevel", "l", "", "set log level")
	fs.StringVarP(&o.LogFileName, "logfile", "L", "", "set log file")
	fs.StringVarP(&o.LogConfig, "logconfig", "", "", "log config")
	fs.StringArrayVarP(&o.LogKeys, "logkeys", "", nil, "log tags/realms(.) to be enabled ([.]name{,[.]name}[=level])")
}

func (o *Options) Close() error {
	if o.LogFile == nil {
		return nil
	}
	return o.LogFile.Close()
}

func (o *Options) Configure(ctx ocm.Context, logctx logging.Context) error {
	var err error

	if logctx == nil {
		// by default: always configure the root logging context used for the actual ocm context.
		logctx = ctx.LoggingContext()
		for logctx.Tree().GetBaseContext() != nil {
			logctx = logctx.Tree().GetBaseContext()
		}
	}

	if o.LogLevel != "" {
		l, err := logging.ParseLevel(o.LogLevel)
		if err != nil {
			return errors.Wrapf(err, "invalid log level %q", o.LogLevel)
		}
		logctx.SetDefaultLevel(l)
	} else {
		logctx.SetDefaultLevel(logging.ErrorLevel)
	}
	logcfg := &config.Config{DefaultLevel: logging.LevelName(logctx.GetDefaultLevel())}

	fs := vfsattr.Get(ctx)
	if o.LogFileName != "" {
		o.LogFile, err = fs.OpenFile(o.LogFileName, vfs.O_CREATE|vfs.O_WRONLY, 0o600)
		if err != nil {
			return errors.Wrapf(err, "cannot open log file %q", o.LogFile)
		}
		log := logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
		log.SetOutput(o.LogFile)
		logctx.SetBaseLogger(logrusr.New(log))
	}

	if o.LogConfig != "" {
		cfg, err := vfs.ReadFile(fs, o.LogConfig)
		if err != nil {
			return errors.Wrapf(err, "cannot read logging config %q", o.LogFile)
		}
		if err = config.ConfigureWithData(logctx, cfg); err != nil {
			return errors.Wrapf(err, "invalid logging config: %q", o.LogFile)
		}
	}

	for _, t := range o.LogKeys {
		level := logging.InfoLevel
		i := strings.Index(t, "=")
		if i >= 0 {
			level, err = logging.ParseLevel(t[i+1:])
			if err != nil {
				return errors.Wrapf(err, "invalid log tag setting")
			}
			t = t[:i]
		}
		var cond []logging.Condition
		cfgcond := []json.RawMessage{}

		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if strings.HasPrefix(tag, "/") {
				cond = append(cond, logging.NewRealmPrefix(tag[1:]))
				data, err := elem("realmprefix", tag[1:])
				if err == nil {
					cfgcond = append(cfgcond, data)
				}
			} else {
				cond = append(cond, logging.NewTag(tag))
				data, err := elem("tag", tag)
				if err == nil {
					cfgcond = append(cfgcond, data)
				}
			}
		}
		rule := logging.NewConditionRule(level, cond...)
		cfgrule := config.ConditionalRule{
			Level:      logging.LevelName(level),
			Conditions: cfgcond,
		}
		data, err := elem("rule", cfgrule)
		if err == nil {
			logcfg.Rules = append(logcfg.Rules, data)
		}
		logctx.AddRule(rule)
	}
	o.LogForward = logcfg
	logforward.Set(ctx.AttributesContext(), logcfg)

	return nil
}

func elem(typ string, obj interface{}) (json.RawMessage, error) {
	return json.Marshal(map[string]interface{}{typ: obj})
}
