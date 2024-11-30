/*
 * Copyright 2022 Mandelsoft. All rights reserved.
 *  This file is licensed under the Apache Software License, v. 2 except as noted
 *  otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package cwdfs

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

type WorkingDirectoryFileSystem struct {
	base vfs.FileSystem
	vol  string
	cwd  string
}

func New(base vfs.FileSystem, path string) (vfs.FileSystemWithWorkingDirectory, error) {
	real, err := vfs.Canonical(base, path, true)
	if err != nil {
		return nil, err
	}
	dir, err := base.Stat(real)
	if err != nil {
		return nil, err
	}
	if !dir.IsDir() {
		return nil, &os.PathError{Op: "readdir", Path: path, Err: errors.New("not a dir")}
	}
	if old, ok := base.(*WorkingDirectoryFileSystem); ok {
		base = old.base
	}
	return &WorkingDirectoryFileSystem{base, base.VolumeName(real), real}, nil
}

func (w *WorkingDirectoryFileSystem) Chdir(path string) error {
	real, err := vfs.Canonical(w, path, true)
	if err != nil {
		return err
	}
	fi, err := w.Lstat(real)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return &os.PathError{Op: "chdir", Path: path, Err: errors.New("no dir")}
	}
	w.cwd = real
	return nil
}

func (w *WorkingDirectoryFileSystem) Name() string {
	return fmt.Sprintf("%s(%s)", w.base.Name(), w.cwd)
}

func (w *WorkingDirectoryFileSystem) VolumeName(name string) string {
	return w.base.VolumeName(name)
}

func (w *WorkingDirectoryFileSystem) FSTempDir() string {
	return w.base.FSTempDir()
}

func (w *WorkingDirectoryFileSystem) Normalize(path string) string {
	return w.base.Normalize(path)
}

func (w *WorkingDirectoryFileSystem) Getwd() (string, error) {
	return w.cwd, nil
}

func (w *WorkingDirectoryFileSystem) realPath(path string) (string, error) {
	vol, path := vfs.SplitVolume(w.base, path)

	if vol != w.vol {
		return "", fmt.Errorf("volume mismatch")
	}
	if vfs.IsAbs(w, path) {
		return vol + path, nil
	}
	return vfs.Join(w.base, w.cwd, path), nil
}

func (w *WorkingDirectoryFileSystem) Create(name string) (vfs.File, error) {
	abs, err := w.realPath(name)
	if err != nil {
		return nil, err
	}
	return w.base.Create(abs)
}

func (w *WorkingDirectoryFileSystem) Mkdir(name string, perm os.FileMode) error {
	abs, err := w.realPath(name)
	if err != nil {
		return err
	}
	return w.base.Mkdir(abs, perm)
}

func (w *WorkingDirectoryFileSystem) MkdirAll(path string, perm os.FileMode) error {
	abs, err := w.realPath(path)
	if err != nil {
		return err
	}
	return w.base.MkdirAll(abs, perm)
}

func (w *WorkingDirectoryFileSystem) Open(name string) (vfs.File, error) {
	abs, err := w.realPath(name)
	if err != nil {
		return nil, err
	}
	return w.base.Open(abs)
}

func (w *WorkingDirectoryFileSystem) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	abs, err := w.realPath(name)
	if err != nil {
		return nil, err
	}
	return w.base.OpenFile(abs, flag, perm)
}

func (w *WorkingDirectoryFileSystem) Remove(name string) error {
	abs, err := w.realPath(name)
	if err != nil {
		return err
	}
	return w.base.Remove(abs)
}

func (w *WorkingDirectoryFileSystem) RemoveAll(path string) error {
	abs, err := w.realPath(path)
	if err != nil {
		return err
	}
	return w.base.RemoveAll(abs)
}

func (w *WorkingDirectoryFileSystem) Rename(oldname, newname string) error {
	absnew, err := w.realPath(newname)
	if err != nil {
		return err
	}
	absold, err := w.realPath(oldname)
	if err != nil {
		return err
	}
	return w.base.Rename(absold, absnew)
}

func (w *WorkingDirectoryFileSystem) Stat(name string) (os.FileInfo, error) {
	abs, err := w.realPath(name)
	if err != nil {
		return nil, err
	}
	return w.base.Stat(abs)
}

func (w *WorkingDirectoryFileSystem) Chmod(name string, mode os.FileMode) error {
	abs, err := w.realPath(name)
	if err != nil {
		return err
	}
	return w.base.Chmod(abs, mode)
}

func (w *WorkingDirectoryFileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	abs, err := w.realPath(name)
	if err != nil {
		return err
	}
	return w.base.Chtimes(abs, atime, mtime)
}

func (w *WorkingDirectoryFileSystem) Lstat(name string) (os.FileInfo, error) {
	abs, err := w.realPath(name)
	if err != nil {
		return nil, err
	}
	return w.base.Lstat(abs)
}

func (w *WorkingDirectoryFileSystem) Symlink(oldname, newname string) error {
	abs, err := w.realPath(newname)
	if err != nil {
		return err
	}
	return w.base.Symlink(oldname, abs)
}

func (w *WorkingDirectoryFileSystem) Readlink(name string) (string, error) {
	abs, err := w.realPath(name)
	if err != nil {
		return "", err
	}
	return w.base.Readlink(abs)
}
