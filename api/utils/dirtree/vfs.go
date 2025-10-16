package dirtree

import (
	"fmt"
	"os"
	"path"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/utils"
)

func NewVFSDirNode(ctx Context, p string, fss ...vfs.FileSystem) (*DirNode, error) {
	fs := utils.FileSystem(fss...)
	f, err := fs.Stat(p)
	if err != nil {
		return nil, err
	}
	if !f.IsDir() {
		return nil, fmt.Errorf("no directory")
	}
	entries, err := vfs.ReadDir(fs, p)
	if err != nil {
		return nil, err
	}
	d := NewDirNode(ctx)

	for _, e := range entries {
		var n Node
		if e.IsDir() {
			n, err = NewVFSDirNode(ctx, path.Join(p, e.Name()), fs)
		} else {
			n, err = NewVFSFileNode(ctx, path.Join(p, e.Name()), fs)
		}
		if err != nil {
			return nil, err
		}
		err = d.AddNode(path.Base(e.Name()), n)
		if err != nil {
			return nil, err
		}
	}
	d.Complete()
	return d, nil
}

func NewVFSFileNode(ctx Context, p string, fss ...vfs.FileSystem) (Node, error) {
	fs := utils.FileSystem(fss...)

	fi, err := fs.Lstat(p)
	if err != nil {
		return nil, err
	}

	t := fi.Mode() & os.ModeType
	if t != 0 && t != os.ModeSymlink {
		return nil, errors.ErrNotSupported("filetype", fmt.Sprintf("%o", t))
	}
	if t == os.ModeSymlink {
		l, err := fs.Readlink(p)
		if err != nil {
			return nil, err
		}
		return NewLinkNode(ctx, l)
	}

	f, err := fs.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewFileNode(ctx, fi.Mode(), fi.Size(), f)
}
