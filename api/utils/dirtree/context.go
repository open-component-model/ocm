package dirtree

import (
	"crypto/sha1" //nolint:gosec // required
	"fmt"
	"hash"
	"io"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils"
)

type Context interface {
	logging.Context
	Hasher() hash.Hash
	FileMode(vfs.FileMode) Mode
	DirMode(vfs.FileMode) Mode
	LinkMode(vfs.FileMode) Mode
	WriteHeader(w io.Writer, typ string, size int64) error
}

// DefaultContext provides a default directory tree hashing context.
// It is based on the Git tree hash mechanism.
func DefaultContext(ctx ...logging.Context) Context {
	return &defaultContext{utils.OptionalDefaulted(LogContext, ctx...)}
}

type defaultContext struct {
	logging.Context
}

func (d defaultContext) Hasher() hash.Hash {
	return sha1.New() //nolint:gosec // required
}

func (d defaultContext) FileMode(mode vfs.FileMode) Mode {
	return FileMode(mode) | ModeBlob
}

func (d defaultContext) DirMode(mode vfs.FileMode) Mode {
	return ModeDir
}

func (d defaultContext) LinkMode(mode vfs.FileMode) Mode {
	return ModeSym
}

func (d defaultContext) WriteHeader(w io.Writer, typ string, size int64) error {
	_, err := w.Write([]byte(fmt.Sprintf("%s %d\000", typ, size)))
	return err
}
