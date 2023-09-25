// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compose

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const Name = "compose"

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <name> <options json> <base spec json>",
		Short: "compose access specification from options and base specification",
		Long: `
The task of this command is to compose an access specification based on some
explicitly given input options and preconfigured specifications.

The finally composed access specification has to be returned as JSON document
on *stdout*.

This command is only used, if for an access method descriptor configuration
options are defined (<CMD>` + p.Name() + ` descriptor</CMD>).

If possible, predefined standard options should be used. In such a case only the
<code>name</code> field should be defined for an option. If required, new options can be
defined by additionally specifying a type and a description. New options should
be used very carefully. The chosen names MUST not conflict with names provided
by other plugins. Therefore, it is highly recommended to use use names prefixed
by the plugin name.

` + options.DefaultRegistry.Usage(),
		Args: cobra.ExactArgs(3),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Command(p, cmd, &opts)
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}

type Options struct {
	Name    string
	Options ppi.Config
	Base    ppi.Config
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[1]), &o.Options); err != nil {
		return errors.Wrapf(err, "invalid access specification options")
	}
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[2]), &o.Base); err != nil {
		return errors.Wrapf(err, "invalid base access specification")
	}
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	k, v := runtime.KindVersion(opts.Name)
	m := p.GetAccessMethod(k, v)
	if m == nil {
		return errors.ErrUnknown(errors.KIND_ACCESSMETHOD, opts.Name)
	}
	err := opts.Options.ConvertFor(m.Options()...)
	if err != nil {
		return err
	}
	err = m.ComposeAccessSpecification(p, opts.Options, opts.Base)
	if err != nil {
		return err
	}
	data, err := json.Marshal(opts.Base)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
