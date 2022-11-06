// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/common"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	Name     = "download"
	OptMedia = common.OptMedia
	OptArt   = common.OptArt
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] <name> <filepath>",
		Short: "download blob into filesystem",
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
	Name string
	Path string

	MediaType    string
	ArtifactType string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.MediaType, OptMedia, "m", "", "media type of input blob")
	fs.StringVarP(&o.ArtifactType, OptArt, "a", "", "artifact type of input blob")
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	o.Path = args[1]
	return nil
}

type Result struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	d := p.GetDownloader(opts.Name)
	if d == nil {
		return errors.ErrNotFound(ppi.KIND_DOWNLOADER, fmt.Sprintf("%s:%s", opts.ArtifactType, opts.MediaType))
	}
	w, h, err := d.Writer(p, opts.ArtifactType, opts.MediaType, opts.Path)
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
	path, err := h()
	result := Result{
		Path: path,
	}
	if err != nil {
		result.Error = err.Error()
	}
	data, err := json.Marshal(result)
	if err == nil {
		cmd.Printf("%s\n", string(data))
	}
	return err
}
