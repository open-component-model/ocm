// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package put

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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
		Use:   Name + " [<flags>] <name> <repository specification>",
		Short: "upload blob to external repository",
		Long:  "",
		Args:  cobra.ExactArgs(2),
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
	Name          string
	Specification json.RawMessage

	Credentials  credentials.DirectCredentials
	MediaType    string
	ArtifactType string

	Hint string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Credentials, "credentials", "c", nil, "credentials")
	flag.StringToStringVarPF(fs, &o.Credentials, "credential", "C", nil, "dedicated credential value")
	fs.StringVarP(&o.MediaType, "mediaType", "m", "", "media type of input blob")
	fs.StringVarP(&o.ArtifactType, "artifactType", "a", "", "artifact type of input blob")
	fs.StringVarP(&o.Hint, "hint", "H", "", "reference hint for storing blob")
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[1]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid repository specification")
	}
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeUploadTargetSpecification(opts.Specification)
	if err != nil {
		return err
	}

	u := p.GetUploader(opts.Name)
	if u == nil {
		return errors.ErrNotFound(ppi.KIND_UPLOADER, fmt.Sprintf("%s:%s", opts.ArtifactType, opts.MediaType))
	}
	w, h, err := u.Writer(p, opts.ArtifactType, opts.MediaType, opts.Hint, spec, opts.Credentials)
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
	acc := h()
	data, err := json.Marshal(acc)
	if err == nil {
		cmd.Printf("%s\n", string(data))
	}
	return err
}
