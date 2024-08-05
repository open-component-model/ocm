package fileblob

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
)

type Option = optionutils.Option[*Options]

type compressionMode string

const (
	COMPRESSION   = compressionMode("compression")
	DECOMPRESSION = compressionMode("decompression")
	NONE          = compressionMode("")
)

type Options struct {
	api.Options
	FileSystem  vfs.FileSystem
	Compression compressionMode
}

var (
	_ api.GeneralOptionsProvider = (*Options)(nil)
	_ Option                     = (*Options)(nil)
)

func (o *Options) ApplyTo(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	if o.FileSystem != nil {
		opts.FileSystem = o.FileSystem
	}
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return api.WrapHint[Options](h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return api.WrapGlobalAccess[Options](a)
}

////////////////////////////////////////////////////////////////////////////////
// Local Options

type filesystem struct {
	fs vfs.FileSystem
}

func (o filesystem) ApplyTo(opts *Options) {
	opts.FileSystem = o.fs
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return filesystem{fs}
}

////////////////////////////////////////////////////////////////////////////////

type compression struct {
	mode compressionMode
}

func (o compression) ApplyTo(opts *Options) {
	opts.Compression = o.mode
}

func WithCompression() Option {
	return compression{COMPRESSION}
}

func WithDecompression() Option {
	return compression{DECOMPRESSION}
}
