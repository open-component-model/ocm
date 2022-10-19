// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
)

var (
	Names = names.Cache
	Verb  = verbs.Info
)

type Cache interface {
	accessio.BlobCache
	accessio.RootedCache
}

type Command struct {
	utils.BaseCommand
	cache Cache
}

// NewCommand creates a new artefact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "",
		Short: "show OCI blob cache information",
		Long: `
Show details about the OCI blob cache (if given).
	`,
		Args: cobra.NoArgs,
		Example: `
$ ocm cache info
`,
	}
}

func (o *Command) Complete(args []string) error {
	c := cacheattr.Get(o.Context)
	if c == nil {
		return errors.Newf("no blob cache configured")
	}
	r, ok := c.(Cache)
	if !ok {
		return errors.Newf("only filesystem based caches are supported")
	}
	o.cache = r
	return nil
}

func (o *Command) Run() error {
	var size int64
	cnt := 0

	if l, ok := o.cache.(sync.Locker); ok {
		l.Lock()
		defer l.Unlock()
	}
	path, fs := o.cache.Root()

	entries, err := vfs.ReadDir(fs, path)
	if err != nil {
		return err
	}
	for _, e := range entries {
		cnt++
		size += e.Size()
	}

	out.Outf(o.Context, "Used cache directory %s [%s]\n", path, fs.Name())
	out.Outf(o.Context, "Total cache size %d entries [%.2f MB]\n", cnt, float64(size)/1024/1024)
	return nil
}
