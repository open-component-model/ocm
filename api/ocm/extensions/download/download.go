package download

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
)

type Option = optionutils.Option[*Options]

type Options struct {
	Printer    common.Printer
	FileSystem vfs.FileSystem
}

func (o *Options) ApplyTo(opts *Options) {
	if o.Printer != nil {
		opts.Printer = o.Printer
	}
	if o.FileSystem != nil {
		opts.FileSystem = o.FileSystem
	}
}

////////////////////////////////////////////////////////////////////////////////

type filesystem struct {
	fs vfs.FileSystem
}

func (o *filesystem) ApplyTo(opts *Options) {
	if o.fs != nil {
		opts.FileSystem = o.fs
	}
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return &filesystem{fs}
}

////////////////////////////////////////////////////////////////////////////////

type printer struct {
	pr common.Printer
}

func (o *printer) ApplyTo(opts *Options) {
	if o.pr != nil {
		opts.Printer = o.pr
	}
}

func WithPrinter(pr common.Printer) Option {
	return &printer{pr}
}

////////////////////////////////////////////////////////////////////////////////

func DownloadResource(ctx cpi.ContextProvider, r cpi.ResourceAccess, path string, opts ...Option) (string, error) {
	eff := optionutils.EvalOptions(opts...)

	fs := utils.FileSystem(eff.FileSystem)
	pr := utils.OptionalDefaulted(common.NewPrinter(nil), eff.Printer)
	_, tgt, err := For(ctx).Download(pr, r, path, fs)
	return tgt, err
}
