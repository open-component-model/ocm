package config

import (
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/keyoption"
	common2 "github.com/open-component-model/ocm/pkg/clisupport"
	"github.com/open-component-model/ocm/pkg/cobrautils/logopts"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	config2 "github.com/open-component-model/ocm/pkg/contexts/config/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	datacfg "github.com/open-component-model/ocm/pkg/contexts/datacontext/config/attrs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/spf13/pflag"
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

	Config      []string                 `json:"config,omitempty"`
	ConfigSets  []string                 `json:"configSets,omitempty"`
	Credentials []string                 `json:"credentials,omitempty"`
	Settings    []string                 `json:"settings,omitempty"`
	Verbose     bool                     `json:"verbose,omitempty"`
	Signing     keyoption.ConfigFragment `json:"signing,omitempty"`
	Logging     logopts.ConfigFragment   `json:"logging,omitempty"`
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
	fs.StringArrayVarP(&c.Config, "config", "", nil, "configuration file")
	fs.StringSliceVarP(&c.ConfigSets, "config-set", "", nil, "apply configuration set")
	fs.StringArrayVarP(&c.Credentials, "cred", "C", nil, "credential setting")
	fs.StringArrayVarP(&c.Settings, "attribute", "X", nil, "attribute setting")
	fs.BoolVarP(&c.Verbose, "verbose", "v", false, "deprecated: enable logrus verbose logging")

	c.Logging.AddFlags(fs)
	c.Signing.AddFlags(fs)
}

func (c *Config) Evaluate(ctx ocm.Context) (*EvaluatedOptions, error) {
	c.Type = ConfigTypeV1
	cfg, err := config2.NewAggregator(true)
	if err != nil {
		return nil, err
	}

	logopts, err := c.Logging.Evaluate(ctx, logging.Context())
	if err != nil {
		return nil, err
	}
	opts := &EvaluatedOptions{LogOpts: logopts}

	opts.Keys, err = c.Signing.Evaluate(ctx, nil)
	if err != nil {
		return opts, err
	}

	if len(c.Config) == 0 {
		_, eff, err := utils.Configure2(ctx, "", vfsattr.Get(ctx))
		if eff != nil {
			err = cfg.AddConfig(eff)
		}
		if err != nil {
			return opts, err
		}
	}
	for _, config := range c.Config {
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

func (c *Config) ApplyTo(_ config.Context, target interface{}) error {
	ctx, ok := target.(cpi.Context)
	if !ok {
		return config.ErrNoContext(ConfigType)
	}

	opts, err := c.Evaluate(ctx)
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
