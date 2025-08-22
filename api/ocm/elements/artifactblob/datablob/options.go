package datablob

import (
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
)

type Option interface {
	ApplyTo(opts *Options)
}

type OptionFunc func(opts *Options)

func (f OptionFunc) ApplyTo(opts *Options) {
	f(opts)
}

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
	if opts == nil {
		return
	}
	o.Options.ApplyTo(&opts.Options)
	if o.MimeType != "" {
		opts.MimeType = o.MimeType
	}
	if o.Compression != NONE {
		opts.Compression = o.Compression
	}
}

func (o *Options) Apply(opts ...Option) {
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(o)
		}
	}
}

// //////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return OptionFunc(func(opts *Options) {
		api.WithHint(h).ApplyTo(&opts.Options)
	})
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return OptionFunc(func(opts *Options) {
		api.WithGlobalAccess(a).ApplyTo(&opts.Options)
	})
}

// //////////////////////////////////////////////////////////////////////////////
// Local Options

func WithMimeType(mime string) Option {
	return OptionFunc(func(opts *Options) {
		opts.MimeType = mime
	})
}

func WithCompression() Option {
	return OptionFunc(func(opts *Options) {
		opts.Compression = COMPRESSION
	})
}

func WithDecompression() Option {
	return OptionFunc(func(opts *Options) {
		opts.Compression = DECOMPRESSION
	})
}
