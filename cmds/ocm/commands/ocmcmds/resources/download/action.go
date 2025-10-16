package download

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/utils/blobaccess"
	common2 "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/common/options/destoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/storeoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/output"
)

////////////////////////////////////////////////////////////////////////////////

type Action struct {
	downloaders download.Registry
	data        elemhdlr.Objects
	opts        *output.Options
}

func NewAction(ctx ocm.ContextProvider, opts *output.Options) *Action {
	return &Action{downloaders: download.For(ctx), opts: opts}
}

func (d *Action) AddOptions(opts ...options.Options) {
	d.opts.OptionSet = append(d.opts.OptionSet, opts...)
}

func (d *Action) Add(e interface{}) error {
	d.data = append(d.data, e.(*elemhdlr.Object))
	return nil
}

func (d *Action) Close() error {
	if len(d.data) == 0 {
		out.Outf(d.opts.Context, "no resources selected\n")
	}
	return nil
}

func (d *Action) Out() error {
	list := errors.ErrListf("downloading resources")
	dest := destoption.From(d.opts)
	if len(d.data) == 1 {
		if dest.Destination == "" {
			_, _ = common.Elem(d.data[0]).Labels.GetValue("downloadName", &dest.Destination)
		}
		return d.Save(d.data[0], dest.Destination)
	} else {
		if dest.Destination == "-" {
			return fmt.Errorf("standard output supported for single resource only.")
		}
		for _, e := range d.data {
			f := dest.Destination
			if f == "" {
				f = "."
			}
			for _, p := range e.History {
				f = path.Join(f, p.GetName(), p.GetVersion())
			}
			r := common.Elem(e)
			n := ""
			if ok, err := r.Labels.GetValue("downloadName", &n); !ok || err != nil {
				n = r.Name
			}
			f = path.Join(f, n)
			id := r.GetIdentity(e.Version.GetDescriptor().Resources)
			delete(id, v1.SystemIdentityName)
			if len(id) > 0 {
				f += "-" + strings.ReplaceAll(id.String(), "\"", "")
			}
			err := d.Save(e, f)
			if err != nil {
				list.Add(err)
				out.Outf(d.opts.Context, "%s failed: %s\n", f, err)
			}
		}
	}
	return list.Result()
}

func (d *Action) Save(o *elemhdlr.Object, f string) error {
	printer := common2.NewPrinter(d.opts.Context.StdOut())
	dest := destoption.From(d.opts)
	local := From(d.opts)
	verify := storeoption.From(d.opts)
	pathIn := true
	r := common.Elem(o)
	if f == "" {
		pathIn = false
	}
	if verify.Store != nil {
		local.Verify = true
	}

	var tmp vfs.File
	var err error
	if f == "-" {
		tmp, err = vfs.TempFile(dest.PathFilesystem, "", "download-*")
		if err != nil {
			return err
		}
		f = tmp.Name()
		tmp.Close()
		printer = common2.NewPrinter(nil)
		defer dest.PathFilesystem.Remove(f)
	}
	id := r.GetIdentity(o.Version.GetDescriptor().Resources)
	racc, err := o.Version.GetResource(id)
	if err != nil {
		return err
	}
	dir := path.Dir(f)
	if dir != "" && dir != "." {
		err = dest.PathFilesystem.MkdirAll(dir, 0o770)
		if err != nil {
			return err
		}
	}
	var ok bool
	var eff string
	if local.UseHandlers {
		ok, eff, err = d.downloaders.Download(printer, racc, f, dest.PathFilesystem)
	} else {
		ok, eff, err = d.downloaders.DownloadAsBlob(printer, racc, f, dest.PathFilesystem)

		if local.Verify {
			var done bool
			done, err = signing.VerifyResourceDigest(o.Version, o.Index, blobaccess.DataAccessForFile(dest.PathFilesystem, f), verify.Store)
			if err != nil {
				if done {
					printer.Printf("%s: verification failed: %s\n", eff, err)
				} else {
					printer.Printf("%s: cannot verify: %s\n", eff, err)
				}
			} else {
				if done {
					printer.Printf("%s: resource content verified\n", eff)
				} else {
					printer.Printf("%s: no resource content verification possible\n", eff)
				}
			}
		}
	}
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newf("no downloader configured for type %q", racc.Meta().GetType())
	}
	if tmp != nil {
		if eff != f {
			defer dest.PathFilesystem.Remove(eff)
		}
		file, err := dest.PathFilesystem.Open(eff)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(d.opts.Context.StdOut(), file)
		if err != nil {
			return err
		}
	} else if eff != f && pathIn {
		out.Outf(d.opts.Context, "output path %q changed to %q by downloader", f, eff)
	}
	return nil
}
