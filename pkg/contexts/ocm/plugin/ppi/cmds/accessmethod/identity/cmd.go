// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/v2/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/v2/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/ppi"
	commonppi "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/ppi/cmds/common"
	"github.com/open-component-model/ocm/v2/pkg/errors"
	"github.com/open-component-model/ocm/v2/pkg/runtime"
)

const (
	Name     = "identity"
	OptCreds = commonppi.OptCreds
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] <access spec>",
		Short: "get blob identity",
		Long: `
Evaluate the given access specification and return a inexpensive identity of the blob content if possible on
*stdout*.`,
		Args: cobra.ExactArgs(1),
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
	Credentials   credentials.DirectCredentials
	Specification json.RawMessage
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Credentials, OptCreds, "c", nil, "credentials")
	flag.StringToStringVarPFA(fs, &o.Credentials, "credential", "C", nil, "dedicated credential value")
}

func (o *Options) Complete(args []string) error {
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[0]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid repository specification")
	}
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeAccessSpecification(opts.Specification)
	if err != nil {
		return errors.Wrapf(err, "access specification")
	}

	m := p.GetAccessMethod(runtime.KindVersion(spec.GetType()))
	if m == nil {
		return errors.ErrUnknown(descriptor.KIND_ACCESSMETHOD, spec.GetType())
	}

	_, err = m.ValidateSpecification(p, spec)
	if err != nil {
		return err
	}

	idp, ok := m.(ppi.ContentVersionIdentityProvider)
	if !ok {
		fmt.Println("")
		return nil
	}

	id, err := idp.GetInexpensiveContentVersionIdentity(p, spec, opts.Credentials)
	if err != nil {
		return err
	}
	fmt.Println(id)
	return err
}
