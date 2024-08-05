package datablob

import (
	"github.com/mandelsoft/goutils/optionutils"

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
	MimeType    string
	Compression compressionMode
}

var (
	_ api.GeneralOptionsProvider = (*Options)(nil)
	_ Option                     = (*Options)(nil)
)

func (o *Options) ApplyTo(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	if o.MimeType != "" {
		opts.MimeType = o.MimeType
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

type mimetype struct {
	mime string
}

func (o mimetype) ApplyTo(opts *Options) {
	opts.MimeType = o.mime
}

func WithMimeType(mime string) Option {
	return mimetype{mime}
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
