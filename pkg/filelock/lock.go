package filelock

import (
	"io"
	"os"
	"sync"

	"github.com/juju/fslock"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const DIRECTORY_LOCK = ".lock"

// Mutex is a lock object based on a file.
// It features a process lock and an
// in-process mutex. Therefore, it can be used
// in a GO program in multiple Go-routines to achieve a
// global synchronization among multiple processes
// on the same machine.
// The result of a Lock operation is an io.Closer
// which is used to release the lock again.
type Mutex struct {
	lock     sync.Mutex
	path     string
	lockfile *fslock.Lock
}

func (m *Mutex) Lock() (io.Closer, error) {
	m.lock.Lock()
	if m.lockfile == nil {
		m.lockfile = fslock.New(m.path)
	}
	err := m.lockfile.Lock()
	if err != nil {
		m.lock.Unlock()
		return nil, err
	}
	return &lock{mutex: m}, nil
}

func (m *Mutex) TryLock() (io.Closer, error) {
	if !m.lock.TryLock() {
		return nil, nil
	}
	if m.lockfile == nil {
		m.lockfile = fslock.New(m.path)
	}
	err := m.lockfile.TryLock()
	if err != nil {
		m.lock.Unlock()
		if errors.Is(err, fslock.ErrLocked) {
			err = nil
		}
		return nil, err
	}
	return &lock{mutex: m}, nil
}

func (m *Mutex) Path() string {
	return m.path
}

type lock struct {
	lock  sync.Mutex
	mutex *Mutex
}

func (l *lock) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.mutex == nil {
		return os.ErrClosed
	}
	l.mutex.lockfile.Unlock()
	l.mutex.lock.Unlock()
	l.mutex = nil
	return nil
}

var (
	_filelocks = map[string]*Mutex{}
	_lock      sync.Mutex
)

// MutexFor provides a canonical lock
// for the given file. If the file path describes
// a directory, the lock will be the file .lock
// inside this directory.
func MutexFor(path string) (*Mutex, error) {
	ok, err := vfs.Exists(osfs.OsFs, path)
	if err != nil {
		return nil, err
	}

	var file string
	if ok {
		file, err = vfs.Canonical(osfs.OsFs, path, true)
		if err != nil {
			return nil, err
		}
		ok, err = vfs.IsDir(osfs.OsFs, path)
		if ok {
			file = filepath.Join(file, DIRECTORY_LOCK)
		}
	} else {
		// canonical path is canonical path of directory plus base name of path
		dir := filepath.Dir(path)
		dir, err = vfs.Canonical(osfs.OsFs, dir, true)
		if err == nil {
			file = filepath.Join(dir, filepath.Base(path))
		}
	}
	if err != nil {
		return nil, err
	}

	_lock.Lock()
	defer _lock.Unlock()

	mutex := _filelocks[file]
	if mutex == nil {
		mutex = &Mutex{
			path: file,
		}
		_filelocks[file] = mutex
	}
	return mutex, nil
}

func LockDir(dir string) (io.Closer, error) {
	m, err := MutexFor(filepath.Join(dir, ".lock"))
	if err != nil {
		return nil, err
	}
	return m.Lock()
}

func Lock(path string) (io.Closer, error) {
	m, err := MutexFor(path)
	if err != nil {
		return nil, err
	}
	return m.Lock()
}
