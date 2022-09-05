/*
Copyright (c) for portions of walk.go are held by The Go Authors, 2009 and are
provided under the BSD license.

https://github.com/golang/go/blob/master/LICENSE

Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sympath

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"
)

// Walk walks the file tree rooted at root, calling walkFn for each file or directory
// in the tree, including root. All errors that arise visiting files and directories
// are filtered by walkFn. The files are walked in lexical order, which makes the
// output deterministic but means that for very large directories Walk can be
// inefficient. Walk follows symbolic links.
func Walk(fs vfs.FileSystem, root string, walkFn vfs.WalkFunc) error {
	info, err := fs.Lstat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = symwalk(fs, root, info, walkFn)
	}

	if err != nil && errors.Is(err, filepath.SkipDir) {
		return fmt.Errorf("failed to walk fs: %w", err)
	}

	return nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(fs vfs.FileSystem, dirname string) ([]string, error) {
	f, err := fs.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

// symwalk recursively descends path, calling walkFn.
func symwalk(fs vfs.FileSystem, path string, info os.FileInfo, walkFn vfs.WalkFunc) error {
	// Recursively walk symlinked directories.
	if IsSymlink(info) {
		resolved, err := vfs.EvalSymlinks(fs, path)
		if err != nil {
			return errors.Wrapf(err, "error evaluating symlink %s", path)
		}

		log.Printf("found symbolic link in path: %s resolves to %s", path, resolved)

		if info, err = fs.Lstat(resolved); err != nil {
			return fmt.Errorf("failed to fetch lstat: %w", err)
		}

		if err := symwalk(fs, path, info, walkFn); err != nil && errors.Is(err, filepath.SkipDir) {
			return fmt.Errorf("error walking on symlink: %w", err)
		}

		return nil
	}

	if err := walkFn(path, info, nil); err != nil {
		return fmt.Errorf("failed to walk with function: %w", err)
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readDirNames(fs, path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := filepath.Join(path, name)

		fileInfo, err := fs.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && errors.Is(err, filepath.SkipDir) {
				return fmt.Errorf("failed to walk with function: %w", err)
			}
		} else {
			err = symwalk(fs, filename, fileInfo, walkFn)
			if err != nil {
				if (!fileInfo.IsDir() && !IsSymlink(fileInfo)) || errors.Is(err, filepath.SkipDir) {
					return fmt.Errorf("error walking on symlink: %w", err)
				}
			}
		}
	}

	return nil
}

// IsSymlink is used to determine if the fileinfo is a symbolic link.
func IsSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}
