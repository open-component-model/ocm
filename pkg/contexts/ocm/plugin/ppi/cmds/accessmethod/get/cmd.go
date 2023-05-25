// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	commonppi "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/common"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	Name     = "get"
	OptCreds = commonppi.OptCreds
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] <access spec>",
		Short: "get blob",
		Long: `
Evaluate the given access specification and return the describes blob on
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

	fmt.Fprintf(os.Stderr, "credentials: %s\n", o.Credentials.String())
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
	r, err := m.Reader(p, spec, opts.Credentials)
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, r)
	r.Close()
	return err
}
