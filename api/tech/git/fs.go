package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/go-git/go-billy/v5"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func VFSBillyFS(fsToWrap vfs.FileSystem) (billy.Filesystem, error) {
	if fsToWrap == nil {
		fsToWrap = vfs.New(memoryfs.New())
	}
	fi, err := fsToWrap.Stat(".")
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("invalid vfs for billy conversion: %w", err)
	}

	return &fs{
		FileSystem: fsToWrap,
	}, nil
}

type fs struct {
	vfs.FileSystem
}

var _ billy.Filesystem = &fs{}

// file is a wrapper around a vfs.File that implements billy.File.
// it uses a mutex to lock the file, so it can be used concurrently from the same process, but
// not across processes (like a flock).
type file struct {
	vfs.File
	lockMu sync.Mutex
}

var _ billy.File = &file{}

func (f *file) Lock() error {
	f.lockMu.Lock()
	return nil
}

func (f *file) Unlock() error {
	f.lockMu.Unlock()
	return nil
}

var _ billy.File = &file{}

func (f *fs) Create(filename string) (billy.File, error) {
	vfsFile, err := f.FileSystem.Create(filename)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile)
}

// vfsToBillyFileInfo converts a vfs.File to a billy.File
func (f *fs) vfsToBillyFileInfo(vf vfs.File) (billy.File, error) {
	return &file{
		File: vf,
	}, nil
}

func (f *fs) Open(filename string) (billy.File, error) {
	vfsFile, err := f.FileSystem.Open(filename)
	if errors.Is(err, syscall.ENOENT) {
		return nil, os.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile)
}

func (f *fs) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	if flag&os.O_CREATE != 0 {
		if err := f.FileSystem.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			return nil, err
		}
	}
	vfsFile, err := f.FileSystem.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile)
}

func (f *fs) Stat(filename string) (os.FileInfo, error) {
	fi, err := f.FileSystem.Stat(filename)
	if errors.Is(err, syscall.ENOENT) {
		return nil, os.ErrNotExist
	}
	return fi, err
}

func (f *fs) Rename(oldpath, newpath string) error {
	dir := filepath.Dir(newpath)
	if dir != "." {
		if err := f.FileSystem.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return f.FileSystem.Rename(oldpath, newpath)
}

func (f *fs) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (f *fs) TempFile(dir, prefix string) (billy.File, error) {
	vfsFile, err := vfs.TempFile(f.FileSystem, dir, prefix)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile)
}

func (f *fs) ReadDir(path string) ([]os.FileInfo, error) {
	return vfs.ReadDir(f.FileSystem, path)
}

func (f *fs) Lstat(filename string) (os.FileInfo, error) {
	fi, err := f.FileSystem.Lstat(filename)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) {
			return nil, os.ErrNotExist
		}
	}
	return fi, err
}

func (f *fs) Chroot(path string) (billy.Filesystem, error) {
	fi, err := f.FileSystem.Stat(path)
	if os.IsNotExist(err) {
		if err = f.FileSystem.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
		fi, err = f.FileSystem.Stat(path)
	}

	if err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", path)
	}

	chfs, err := projectionfs.New(f.FileSystem, path)
	if err != nil {
		return nil, err
	}

	return &fs{
		FileSystem: chfs,
	}, nil
}

func (f *fs) Root() string {
	if root := projectionfs.Root(f.FileSystem); root != "" {
		return root
	}
	if canonicalRoot, err := vfs.Canonical(f.FileSystem, "/", true); err == nil {
		return canonicalRoot
	}
	return "/"
}

var _ billy.Filesystem = &fs{}
