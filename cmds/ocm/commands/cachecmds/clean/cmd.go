// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package clean

import (
	"fmt"
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/pkg/out"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
	"github.com/open-component-model/ocm/pkg/errors"

	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/names"
)

var (
	Names = names.Cache
	Verb  = verbs.Clean
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
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "",
		Short: "cleanup oci blob cache",
		Long: `
Cleanup all blobs stored in oci blob cache (if given).
	`,
		Args: cobra.NoArgs,
		Example: `
$ ocm clean cache
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
	var fsize int64
	cnt := 0
	errs := 0

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
		err := fs.RemoveAll(vfs.Join(fs, path, e.Name()))
		if err != nil {
			out.Errf(o.Context, "cannot delete %q: %s\n", e.Name(), err)
			errs++
			fsize += e.Size()
		} else {
			cnt++
			size += e.Size()
		}
	}
	if cnt == 0 && errs > 0 {
		return fmt.Errorf("Failed to delete %d entries [%.2f MB]\n", cnt, float64(fsize)/1024/1024)
	}
	if errs == 0 {
		out.Outf(o.Context, "Successfully deleted %d entries [%.2f MB]\n", cnt, float64(size)/1024/1024)
	} else {
		out.Outf(o.Context, "Deleted %d entries [%.2f MB]\n", cnt, float64(size)/1024/1024)
		out.Outf(o.Context, "Failed to delete %d entries [%.2f MB]\n", cnt, float64(fsize)/1024/1024)
	}
	return nil
}
