// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"fmt"
	"io"
	"strings"

	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/destoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/handlers/artifacthdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers/dirtree"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

var (
	Names = names.Artifacts
	Verb  = verbs.Download
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new download command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), output.OutputOptions(outputs, New(), destoption.New(), &formatoption.Option{}))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  {<artifact>} ",
		Args:  cobra.MinimumNArgs(1),
		Short: "download oci artifacts",
		Long: `
Download artifacts from an OCI registry. The result is stored in
artifact set format, without the repository part

The files are named according to the artifact repository name.
`,
	}
}

func (o *Command) Complete(args []string) error {
	var err error
	o.Refs = args
	return err
}

func (o *Command) Run() error {
	session := oci.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(common.CompleteOptionsWithContext(o.Context, session))
	if err != nil {
		return err
	}

	hdlr := artifacthdlr.NewTypeHandler(o.Context.OCI(), session, repooption.From(o).Repository)
	return utils.HandleArgs(output.From(o), hdlr, o.Refs...)
}

////////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getDownload)

func getDownload(opts *output.Options) output.Output {
	return &download{opts: opts}
}

type download struct {
	data artifacthdlr.Objects
	opts *output.Options
}

func (d *download) Add(e interface{}) error {
	d.data = append(d.data, e.(*artifacthdlr.Object))
	return nil
}

func (d *download) Close() error {
	return nil
}

func (d *download) Out() error {
	list := errors.ErrListf("downloading artifacts")
	dest := destoption.From(d.opts)
	if len(d.data) == 0 {
		out.Outf(d.opts.Context, "no artifacts found\n")
	}
	if len(d.data) == 1 {
		f := dest.Destination
		e := d.data[0]
		if f == "" {
			f = composePath(dest, e)
		}
		return d.Save(e, f)
	} else {
		for _, e := range d.data {
			f := composePath(dest, e)
			err := d.Save(e, f)
			if err != nil {
				list.Add(err)
				out.Outf(d.opts.Context, "%s failed: %s\n", f, err)
			}
		}
	}
	return list.Result()
}

func composePath(dest *destoption.Option, e *artifacthdlr.Object) string {
	f := e.Spec.UniformRepositorySpec.String()
	f = strings.ReplaceAll(f, "::", "-")
	f = vfs.Join(dest.PathFilesystem, f, e.Spec.Repository)
	if dest.Destination != "" {
		return vfs.Join(dest.PathFilesystem, dest.Destination, f)
	}
	return f
}

func (d *download) Save(o *artifacthdlr.Object, f string) error {
	opts := From(d.opts)
	dest := destoption.From(d.opts)
	art := o.Artifact
	dir := vfs.Dir(dest.PathFilesystem, f)

	err := dest.PathFilesystem.MkdirAll(dir, 0o770)
	if err != nil {
		return err
	}

	if len(opts.Layers) > 0 {
		var finalize finalizer.Finalizer
		defer finalize.Finalize()

		if !art.IsManifest() {
			return fmt.Errorf("artifact is not manifest artifact to extract layers from")
		}
		layers := art.ManifestAccess().GetDescriptor().Layers
		for _, l := range opts.Layers {
			nested := finalize.Nested()
			if len(layers) <= l {
				return fmt.Errorf("layer %d does not exist", l)
			}
			blob, err := art.GetBlob(layers[l].Digest)
			if err != nil {
				return errors.Wrapf(err, "cannot extract layer %d", l)
			}
			nested.Close(blob)
			r, err := blob.Reader()
			if err != nil {
				return errors.Wrapf(err, "cannot extract layer %d", l)
			}
			nested.Close(r)
			name := f
			if len(opts.Layers) > 1 {
				name = fmt.Sprintf("%s-%d", f, l)
			}
			file, err := dest.PathFilesystem.OpenFile(name, vfs.O_CREATE|vfs.O_TRUNC|vfs.O_WRONLY, 0640)
			if err != nil {
				return errors.Wrapf(err, "cannot create target file %s for layer %d", name, l)
			}
			nested.Close(file)
			n, err := io.Copy(file, r)
			if err != nil {
				return errors.Wrapf(err, "cannot download layer %d to %s", l, name)
			}
			out.Outf(d.opts.Context, "%s: layer %d: %d byte(s) downloaded\n", name, l, n)
			nested.Finalize()
		}
	} else if opts.DirTree {
		format := formatoption.From(d.opts)

		if !art.IsManifest() {
			return fmt.Errorf("no OCI image manifest")
		}
		fs, reader, err := dirtree.New(art.ManifestAccess().GetDescriptor().Config.MediaType).GetForArtifact(art)
		if err != nil {
			return err
		}
		if reader != nil {
			defer reader.Close()
			if format.Format != accessio.FormatDirectory {
				file, err := dest.PathFilesystem.OpenFile(f, vfs.O_CREATE|vfs.O_TRUNC|vfs.O_WRONLY, 0o640)
				if err != nil {
					return err
				}
				defer file.Close()
				written, err := io.Copy(file, reader)
				if err != nil {
					return err
				}
				out.Outf(d.opts.Context, "%s: %d byte(s) downloaded\n", f, written)
				return nil
			} else {
				r, _, err := compression.AutoDecompress(reader)
				if err != nil {
					return errors.Wrapf(err, "cannot determine compression")
				}
				dest.PathFilesystem.MkdirAll(f, 0o740)
				tfs, err := projectionfs.New(dest.PathFilesystem, f)
				if err != nil {
					return err
				}
				fcnt, bcnt, err := tarutils.ExtractTarToFsWithInfo(tfs, r)
				if err != nil {
					return err
				}
				out.Outf(d.opts.Context, "%s: %d files with %d byte(s) downloaded\n", f, fcnt, bcnt)
				return nil
			}
		} else {
			return accessio.CopyFileSystem(format.Format, fs, "/", dest.PathFilesystem, f, 0o640)
		}
	} else {
		blob, err := art.Blob()
		if err != nil {
			return err
		}
		defer blob.Close()

		digest := blob.Digest()
		format := formatoption.From(d.opts)
		set, err := artifactset.Create(accessobj.ACC_CREATE, f, format.Mode(), format.Format, accessio.PathFileSystem(dest.PathFilesystem))
		if err != nil {
			return err
		}
		defer set.Close()

		err = artifactset.TransferArtifact(art, set)
		if err != nil {
			return err
		}

		if o.Spec.Tag != nil {
			err = set.AddTags(digest, *o.Spec.Tag)
			if err != nil {
				return err
			}
		}
		set.Annotate(artifactset.MAINARTIFACT_ANNOTATION, digest.String())
		out.Outf(d.opts.Context, "%s: downloaded\n", f)
	}

	return nil
}
