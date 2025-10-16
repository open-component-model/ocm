package support

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	_ "ocm.software/ocm/api/cli/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	datactg "ocm.software/ocm/api/datacontext/config/attrs"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/tools/toi"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	"ocm.software/ocm/api/utils/clisupport"
	"ocm.software/ocm/api/utils/cobrautils"
	"ocm.software/ocm/api/utils/cobrautils/logopts"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/api/version"
)

type BootstrapperCLIOptions struct {
	ExecutorOptions
	logopts.Options
	CredentialSettings []string
	Settings           []string
}

func NewCLICommand(ctx ocm.Context, name string, exec func(options *ExecutorOptions) error) *cobra.Command {
	if ctx == nil {
		ctx = ocm.DefaultContext()
	}
	opts := &BootstrapperCLIOptions{
		ExecutorOptions: ExecutorOptions{
			Context: ctx,
		},
	}
	cmd := &cobra.Command{
		Use:                   name + " {<options>} [<action>] <component version>",
		Short:                 "Bootstrapper using the OCM bootstrap mechanism",
		Version:               version.Get().String(),
		TraverseChildren:      true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			action := ""
			if len(args) > 0 {
				if len(args) > 1 {
					action = args[0]
					opts.ComponentVersionName = args[1]
				} else {
					opts.ComponentVersionName = args[0]
				}
			}
			opts.Action = action
			return opts.Complete()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out.Outf(opts.OutputContext, "This is %s (%s)\n", name, version.Get().String())
			e := &Executor{Completed: true, Options: &opts.ExecutorOptions, Run: exec}
			return e.Execute()
		},
	}
	cobrautils.TweakCommand(cmd, nil)

	cmd.AddCommand(NewVersionCommand())
	opts.AddFlags(cmd.Flags())
	cmd.InitDefaultHelpCmd()

	return cmd
}

func (o *BootstrapperCLIOptions) AddFlags(fs *pflag.FlagSet) {
	o.Options.AddFlags(fs)
	fs.StringVarP(&o.OCMConfig, "ocmconfig", "", "", "ocm configuration file")
	fs.StringArrayVarP(&o.CredentialSettings, "cred", "C", nil, "credential setting")
	fs.StringArrayVarP(&o.Settings, "attribute", "X", nil, "attribute setting")

	fs.StringVarP(&o.Inputs, "inputs", "", "", "input path")
	fs.StringVarP(&o.Outputs, "outputs", "", "", "output path")
	fs.StringVarP(&o.Root, "bootstraproot", "", install.PathTOI, "bootstrapper contract root folder")
	fs.StringVarP(&o.Config, "config", "", "", "bootstrapper configuration input file")
	fs.StringVarP(&o.Parameters, "parameters", "", "", "bootstrapper parameter input file")
	fs.StringVarP(&o.RepoPath, "ctf", "", "", "bootstrapper transport archive")
}

func (o *BootstrapperCLIOptions) Complete() error {
	o.Options.Configure(o.Context, o.Context.LoggingContext())
	if err := o.ExecutorOptions.Complete(); err != nil {
		return fmt.Errorf("unable to complete options: %w", err)
	}

	id := credentials.ConsumerIdentity{}
	attrs := common.Properties{}

	for _, s := range o.CredentialSettings {
		i := strings.Index(s, "=")
		if i < 0 {
			return errors.ErrInvalid("credential setting", s)
		}

		name := s[:i]
		value := s[i+1:]

		if strings.HasPrefix(name, ":") {
			if len(attrs) != 0 {
				o.Context.CredentialsContext().SetCredentialsForConsumer(id, credentials.NewCredentials(attrs))
				id = credentials.ConsumerIdentity{}
				attrs = common.Properties{}
			}
			name = name[1:]
			id[name] = value
		} else {
			attrs[name] = value
		}

		if len(name) == 0 {
			return errors.ErrInvalid("credential setting", s)
		}
	}

	if len(attrs) != 0 {
		o.Context.CredentialsContext().SetCredentialsForConsumer(id, credentials.NewCredentials(attrs))
	} else if len(id) != 0 {
		return errors.Newf("empty credential attribute set for %s", id.String())
	}

	set, err := clisupport.ParseLabels(vfsattr.Get(o.Context), o.Settings, "attribute setting")
	if err == nil && len(set) > 0 {
		ctx := o.Context.ConfigContext()
		spec := datactg.New()
		for _, s := range set {
			attr := s.Name
			eff := datacontext.DefaultAttributeScheme.Shortcuts()[attr]
			if eff != "" {
				attr = eff
			}
			err = spec.AddRawAttribute(attr, s.Value)
			if err != nil {
				return errors.Wrapf(err, "attribute %s", s.Name)
			}
		}
		err = ctx.ApplyConfig(spec, "cli")
	}

	if err != nil {
		return fmt.Errorf("unable to parse labels: %w", err)
	}

	o.Logger = toi.Log
	return nil
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "displays the version",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			//nolint:forbidigo // It's an intentional Printf.
			fmt.Printf("%#v", v)
		},
	}
}
