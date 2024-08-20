package git

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-billy/v5"
	"github.com/juju/fslock"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func VFSBillyFS(fsToWrap vfs.FileSystem) billy.Filesystem {
	if fsToWrap == nil {
		fsToWrap = vfs.New(memoryfs.New())
	}
	fi, err := fsToWrap.Stat(".")
	if err != nil || !fi.IsDir() {
		panic(fmt.Errorf("invalid vfs: %v", err))
	}

	return &fs{
		vfs: fsToWrap,
	}
}

type fs struct {
	vfs vfs.FileSystem
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
	return f.vfsToBillyFileInfo(vfsFile)
}

// vfsToBillyFileInfo converts a vfs.File to a billy.File
// It also creates a fslock.Lock for the file to ensure that the file is lockable
// If the vfs is an osfs.OsFs, the lock is created in the same directory as the file
// If the vfs is not an osfs.OsFs, a temporary directory is created to store the lock
// because its not trivial to store the lock for jujufs on a virtual filesystem because
// juju vfs only operates on syscalls directly and without interface abstraction its not easy to get the root.
func (f *fs) vfsToBillyFileInfo(vf vfs.File) (billy.File, error) {
	var lock *fslock.Lock
	if f.vfs == osfs.OsFs {
		lock = fslock.New(fmt.Sprintf("%s.lock", vf.Name()))
	} else {
		hash := fnv.New32()
		_, _ = hash.Write([]byte(f.vfs.Name()))
		temp, err := os.MkdirTemp("", fmt.Sprintf("git-vfs-locks-%x", hash.Sum32()))
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir to allow mapping vfs to git (billy) filesystem; "+
				"this temporary directory is mandatory because a virtual filesystem cannot be used to accurately depict os syslocks: %w", err)
		}
		_, components := vfs.Components(f.vfs, vf.Name())
		lockPath := filepath.Join(
			temp,
			filepath.Join(components[:len(components)-1]...),
			fmt.Sprintf("%s.lock", components[len(components)-1]),
		)
		if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create temp dir to allow mapping vfs to git (billy) filesystem; "+
				"this temporary directory is mandatory because a virtual filesystem cannot be used to accurately depict os syslocks: %w", err)
		}
		lock = fslock.New(lockPath)
	}

	return &file{
		vfsFile: vf,
		lock:    lock,
	}, nil
}

func (f *fs) Open(filename string) (billy.File, error) {
	vfsFile, err := f.vfs.Open(filename)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile)
}

func (f *fs) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	if flag&os.O_CREATE != 0 {
		if err := f.vfs.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			return nil, err
		}
	}
	vfsFile, err := f.vfs.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile)
}

func (f *fs) Stat(filename string) (os.FileInfo, error) {
	fi, err := f.vfs.Stat(filename)
	if errors.Is(err, syscall.ENOENT) {
		return nil, os.ErrNotExist
	}
	return fi, err
}

func (f *fs) Rename(oldpath, newpath string) error {
	dir := filepath.Dir(newpath)
	if dir != "." {
		if err := f.vfs.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return f.vfs.Rename(oldpath, newpath)
}

func (f *fs) Remove(filename string) error {
	return f.vfs.Remove(filename)
}

func (f *fs) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (f *fs) TempFile(dir, prefix string) (billy.File, error) {
	vfsFile, err := vfs.TempFile(f.vfs, dir, prefix)
	if err != nil {
		return nil, err
	}
	return f.vfsToBillyFileInfo(vfsFile)
}

func (f *fs) ReadDir(path string) ([]os.FileInfo, error) {
	return vfs.ReadDir(f.vfs, path)
}

func (f *fs) MkdirAll(filename string, perm os.FileMode) error {
	return f.vfs.MkdirAll(filename, perm)
}

func (f *fs) Lstat(filename string) (os.FileInfo, error) {
	fi, err := f.vfs.Lstat(filename)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) {
			return nil, os.ErrNotExist
		}
	}
	return fi, err
}

func (f *fs) Symlink(target, link string) error {
	return f.vfs.Symlink(target, link)
}

func (f *fs) Readlink(link string) (string, error) {
	return f.vfs.Readlink(link)
}

func (f *fs) Chroot(path string) (billy.Filesystem, error) {
	fi, err := f.vfs.Stat(path)
	if os.IsNotExist(err) {
		if err = f.vfs.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
		fi, err = f.vfs.Stat(path)
	}

	if err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", path)
	}

	chfs, err := projectionfs.New(f.vfs, path)
	if err != nil {
		return nil, err
	}

	return &fs{
		vfs: chfs,
	}, nil
}

func (f *fs) Root() string {
	if root := projectionfs.Root(f.vfs); root != "" {
		return root
	}
	if canonicalRoot, err := vfs.Canonical(f.vfs, "/", true); err == nil {
		return canonicalRoot
	}
	return "/"
}

var _ billy.Filesystem = &fs{}
