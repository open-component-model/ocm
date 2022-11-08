// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/common"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	Name     = "validate"
	OptMedia = common.OptMedia
	OptArt   = common.OptArt
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] <name> <spec>",
		Short: "validate upload specification",
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

	ArtifactType string
	MediaType    string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.MediaType, OptMedia, "m", "", "media type of input blob")
	fs.StringVarP(&o.ArtifactType, OptArt, "a", "", "artifact type of input blob")
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[1]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid repository specification")
	}
	return nil
}

type Result struct {
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeUploadTargetSpecification(opts.Specification)
	if err != nil {
		return err
	}

	m := p.GetUploader(opts.Name)
	if m == nil {
		return errors.ErrUnknown(ppi.KIND_UPLOADER, spec.GetType())
	}
	info, err := m.ValidateSpecification(p, spec)
	if err != nil {
		return err
	}
	result := Result{info.ConsumerId}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
