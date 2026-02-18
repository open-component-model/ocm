package config

import (
	"strings"
	"time"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	config2 "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/httptimeoutattr"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	datacfg "ocm.software/ocm/api/datacontext/config/attrs"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/tech/signing"
	common2 "ocm.software/ocm/api/utils/clisupport"
	"ocm.software/ocm/api/utils/cobrautils/logopts"
	logdata "ocm.software/ocm/api/utils/cobrautils/logopts/logging"
	"ocm.software/ocm/api/utils/logging"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/common/options/keyoption"
)

const (
	ConfigType   = "cli.ocm" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based config interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`

	ConfigSets  []string                 `json:"configSets,omitempty"`
	Credentials []string                 `json:"credentials,omitempty"`
	Settings    []string                 `json:"settings,omitempty"`
	Timeout     string                   `json:"timeout,omitempty"`
	Verbose     bool                     `json:"verbose,omitempty"`
	Signing     keyoption.ConfigFragment `json:"signing,omitempty"`
	Logging     logopts.ConfigFragment   `json:"logging,omitempty"`

	// ConfigFiles describes the cli argument for additional config files.
	// This is not persisted since it is resolved by the first evaluation.
	ConfigFiles []string `json:"-"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
	}
}

func (c *Config) GetType() string {
	return ConfigType
}

func (c *Config) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&c.ConfigFiles, "config", "", nil, "configuration file")
	fs.StringSliceVarP(&c.ConfigSets, "config-set", "", nil, "apply configuration set")
	fs.StringArrayVarP(&c.Credentials, "cred", "C", nil, "credential setting")
	fs.StringArrayVarP(&c.Settings, "attribute", "X", nil, "attribute setting")
	fs.StringVar(&c.Timeout, "timeout", "", "client timeout (default 30s, e.g. 30s, 5m)")
	fs.BoolVarP(&c.Verbose, "verbose", "v", false, "deprecated: enable logrus verbose logging")

	c.Logging.AddFlags(fs)
	c.Signing.AddFlags(fs)
}

func (c *Config) Evaluate(ctx ocm.Context, main bool) (*EvaluatedOptions, error) {
	c.Type = ConfigTypeV1
	cfg, err := config2.NewAggregator(true)
	if err != nil {
		return nil, err
	}

	logopts, err := c.Logging.Evaluate(ctx, logging.Context(), main)
	if err != nil {
		return nil, err
	}
	opts := &EvaluatedOptions{LogOpts: logopts}

	opts.Keys, err = c.Signing.Evaluate(ctx, nil)
	if err != nil {
		return opts, err
	}

	if len(c.ConfigFiles) == 0 {
		_, eff, err := utils.Configure2(ctx, "", vfsattr.Get(ctx))
		if eff != nil {
			err = cfg.AddConfig(eff)
		}
		if err != nil {
			return opts, err
		}
	}
	for _, config := range c.ConfigFiles {
		_, eff, err := utils.Configure2(ctx, config, vfsattr.Get(ctx))
		if eff != nil {
			err = cfg.AddConfig(eff)
		}
		if err != nil {
			return opts, err
		}
	}

	keyopts, err := c.Signing.Evaluate(ctx, nil)
	if err != nil {
		return opts, err
	}

	if keyopts.Keys.HasKeys() {
		def := signingattr.Get(ctx)
		err = signingattr.Set(ctx, signing.NewRegistry(def.HandlerRegistry(), signing.NewKeyRegistry(keyopts.Keys, def.KeyRegistry())))
		if err != nil {
			return opts, err
		}
	}

	for _, n := range c.ConfigSets {
		err := ctx.ConfigContext().ApplyConfigSet(n)
		if err != nil {
			return opts, err
		}
	}

	id := credentials.ConsumerIdentity{}
	attrs := common.Properties{}
	for _, s := range c.Credentials {
		i := strings.Index(s, "=")
		if i < 0 {
			return opts, errors.ErrInvalid("credential setting", s)
		}
		name := s[:i]
		value := s[i+1:]
		if strings.HasPrefix(name, ":") {
			if len(attrs) != 0 {
				ctx.CredentialsContext().SetCredentialsForConsumer(id, credentials.NewCredentials(attrs))
				id = credentials.ConsumerIdentity{}
				attrs = common.Properties{}
			}
			name = name[1:]
			id[name] = value
		} else {
			attrs[name] = value
		}
		if len(name) == 0 {
			return opts, errors.ErrInvalid("credential setting", s)
		}
	}
	if len(attrs) != 0 {
		ctx.CredentialsContext().SetCredentialsForConsumer(id, credentials.NewCredentials(attrs))
	} else if len(id) != 0 {
		return opts, errors.Newf("empty credential attribute set for %s", id.String())
	}

	if c.Timeout != "" {
		timeout, err := time.ParseDuration(c.Timeout)
		if err != nil {
			return opts, errors.Wrapf(err, "invalid timeout value %q: use a duration string like 30s, 5m, or 1h", c.Timeout)
		}
		err = ctx.ConfigContext().ApplyConfig(httptimeoutattr.NewConfig(timeout), "cli timeout flag")
		if err != nil {
			return opts, errors.Wrapf(err, "applying timeout config")
		}
	}

	set, err := common2.ParseLabels(vfsattr.Get(ctx), c.Settings, "attribute setting")
	if err != nil {
		return opts, errors.Wrapf(err, "invalid attribute setting")
	}
	if len(set) > 0 {
		cfgctx := ctx.ConfigContext()
		spec := datacfg.New()
		for _, s := range set {
			attr := s.Name
			eff := datacontext.DefaultAttributeScheme.Shortcuts()[attr]
			if eff != "" {
				attr = eff
			}
			err = spec.AddRawAttribute(attr, s.Value)
			if err != nil {
				return opts, errors.Wrapf(err, "attribute %s", s.Name)
			}
		}
		_ = cfgctx.ApplyConfig(spec, "cli")
	}
	cfg.AddConfig(c)
	opts.ConfigForward = cfg.Get()
	return opts, nil
}

func (c *Config) ApplyTo(cctx config.Context, target interface{}) error {
	// first: check for logging config for subsequent command calls
	if lc, ok := target.(*logdata.LoggingConfiguration); ok {
		cfg, err := c.Logging.GetLogConfig(vfsattr.Get(cctx))
		if err != nil {
			return err
		}
		lc.LogConfig = *cfg
		lc.Json = c.Logging.Json
		return nil
	}

	// second: main target is an ocm context
	ctx, ok := target.(cpi.Context)
	if !ok {
		return config.ErrNoContext(ConfigType)
	}

	opts, err := c.Evaluate(ctx, false)
	if err != nil {
		return err
	}
	if opts.Keys.Keys.HasKeys() {
		def := signingattr.Get(ctx)
		err = signingattr.Set(ctx, signing.NewRegistry(def.HandlerRegistry(), signing.NewKeyRegistry(opts.Keys.Keys, def.KeyRegistry())))
		if err != nil {
			return err
		}
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> is used to handle the
main configuration flags of the OCM command line tool.

<pre>
    type: ` + ConfigType + `
    aliases:
       &lt;name>: &lt;OCI registry specification>
       ...
</pre>
`
