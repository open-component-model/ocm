package textblob

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactblob/datablob"
)

type (
	Option  = datablob.Option
	Options = datablob.Options
)

const (
	COMPRESSION   = datablob.COMPRESSION
	DECOMPRESSION = datablob.DECOMPRESSION
	NONE          = datablob.NONE
)

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return datablob.WithHint(h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return datablob.WithGlobalAccess(a)
}

////////////////////////////////////////////////////////////////////////////////
// Local Options

func WithMimeType(mime string) Option {
	return datablob.WithMimeType(mime)
}

func WithCompression() Option {
	return datablob.WithCompression()
}

func WithDecompression() Option {
	return datablob.WithDecompression()
}
