// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package put

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const Name = "put"

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + "<flags> <access spec type>",
		Short: "put blob",
		Long:  "",
		Args:  cobra.ExactArgs(1),
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
	Credentials credentials.DirectCredentials

	MediaType     string
	MethodName    string
	MethodVersion string

	Hint string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Credentials, "credentials", "c", nil, "credentials")
	flag.StringMapVarPA(fs, &o.Credentials, "credential", "C", nil, "dedicated credential value")
	fs.StringVarP(&o.MediaType, "mediaType", "m", "", "media type of input blob")
	fs.StringVarP(&o.Hint, "hint", "h", "", "reference hint for storing blob")
}

func (o *Options) Complete(args []string) error {
	fields := strings.Split(args[0], runtime.VersionSeparator)
	o.MethodName = fields[0]
	if len(fields) > 1 {
		o.MethodVersion = fields[1]
	}
	if len(fields) > 2 {
		return errors.ErrInvalid(errors.KIND_ACCESSMETHOD, args[0])
	}
	return nil
}

func (o *Options) Method() string {
	if o.MethodVersion == "" {
		return o.MethodName
	}
	return o.MethodName + runtime.VersionSeparator + o.MethodVersion
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	m := p.GetAccessMethod(opts.MethodName, opts.MethodVersion)
	if m == nil {
		return errors.ErrUnknown(errors.KIND_ACCESSMETHOD, opts.Method())
	}
	w, h, err := m.Writer(p, opts.MediaType, opts.Credentials)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, os.Stdin)
	if err != nil {
		w.Close()
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	spec := h()
	data, err := json.Marshal(spec)
	if err == nil {
		cmd.Printf("%s\n", string(data))
	}
	return err
}
