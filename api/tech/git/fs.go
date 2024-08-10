package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-billy/v5"
	"github.com/juju/fslock"
	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func VFSBillyFS(fsToWrap vfs.VFS) billy.Filesystem {
	if fsToWrap == nil {
		fsToWrap = vfs.New(memoryfs.New())
	}
	fi, err := fsToWrap.Stat(".")
	if err != nil || !fi.IsDir() {
		panic(fmt.Errorf("invalid vfs: %v", err))
	}

	return &fs{
		vfs:  fsToWrap,
		root: fi.Name(),
	}
}

type fs struct {
	vfs  vfs.VFS
	root string
}

type file struct {
	lock    *fslock.Lock
	vfsFile vfs.File
}

func (f *file) Name() string {
	return f.vfsFile.Name()
}

func (f *file) Write(p []byte) (n int, err error) {
	return f.vfsFile.Write(p)
}

func (f *file) Read(p []byte) (n int, err error) {
	return f.vfsFile.Read(p)
}

func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	return f.vfsFile.ReadAt(p, off)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	return f.vfsFile.Seek(offset, whence)
}

func (f *file) Close() error {
	return f.vfsFile.Close()
}

func (f *file) Lock() error {
	return f.lock.Lock()
}

func (f *file) Unlock() error {
	return f.lock.Unlock()
}

func (f *file) Truncate(size int64) error {
	return f.vfsFile.Truncate(size)
}

var _ billy.File = &file{}

func (f *fs) Create(filename string) (billy.File, error) {
	vfsFile, err := f.vfs.Create(filename)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile), nil
}

func (f *fs) vfsToBillyFileInfo(vf vfs.File) billy.File {
	return &file{
		vfsFile: vf,
		lock:    fslock.New(fmt.Sprintf("%s.lock", vf.Name())),
	}
}

func (f *fs) Open(filename string) (billy.File, error) {
	vfsFile, err := f.vfs.Open(filename)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile), nil
}

func (f *fs) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	vfsFile, err := f.vfs.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile), nil
}

func (f *fs) Stat(filename string) (os.FileInfo, error) {
	fi, err := f.vfs.Stat(filename)
	if errors.Is(err, syscall.ENOENT) {
		return nil, os.ErrNotExist
	}
	return fi, nil
}

func (f *fs) Rename(oldpath, newpath string) error {
	return f.vfs.Rename(oldpath, newpath)
}

func (f *fs) Remove(filename string) error {
	return f.vfs.Remove(filename)
}

func (f *fs) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (f *fs) TempFile(dir, prefix string) (billy.File, error) {
	vfsFile, err := f.vfs.TempFile(dir, prefix)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile), nil
}

func (f *fs) ReadDir(path string) ([]os.FileInfo, error) {
	return f.vfs.ReadDir(path)
}

func (f *fs) MkdirAll(filename string, perm os.FileMode) error {
	return f.vfs.MkdirAll(filename, perm)
}

func (f *fs) Lstat(filename string) (os.FileInfo, error) {
	return f.vfs.Lstat(filename)
}

func (f *fs) Symlink(target, link string) error {
	return f.vfs.Symlink(target, link)
}

func (f *fs) Readlink(link string) (string, error) {
	return f.vfs.Readlink(link)
}

func (f *fs) Chroot(path string) (billy.Filesystem, error) {
	chfs, err := cwdfs.New(f.vfs, path)
	if err != nil {
		return nil, err
	}
	return &fs{
		root: path,
		vfs:  vfs.New(chfs),
	}, nil
}

func (f *fs) Root() string {
	return f.root
}

var _ billy.Filesystem = &fs{}
