package file

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

type Option = optionutils.Option[*Options]

type Options struct {
	// FileSystem defines the file system that contains the specified directory.
	FileSystem vfs.FileSystem
	Digest     digest.Digest
	Size       *int64
}

func (o *Options) GetSize() int64 {
	if o.Size == nil {
		return bpi.BLOB_UNKNOWN_SIZE
	}
	return *o.Size
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.FileSystem != nil {
		opts.FileSystem = o.FileSystem
	}
	if o.Digest != "" {
		opts.Digest = o.Digest
	}
	optionutils.ApplyOption(o.Size, &opts.Size)
}

////////////////////////////////////////////////////////////////////////////////

type fileSystem struct {
	fs vfs.FileSystem
}

func (o *fileSystem) ApplyTo(opts *Options) {
	opts.FileSystem = o.fs
}

func WithFileSystem(fss ...vfs.FileSystem) Option {
	return &fileSystem{fs: utils.FileSystem(fss...)}
}

////////////////////////////////////////////////////////////////////////////////

type size int64

func (o size) ApplyTo(opts *Options) {
	opts.Size = generics.Pointer(int64(o))
}

func WithSize(s int64) Option {
	return size(s)
}

type _digest digest.Digest

func (o _digest) ApplyTo(opts *Options) {
	opts.Digest = digest.Digest(o)
}

func WithDigest(d digest.Digest) Option {
	return _digest(d)
}
