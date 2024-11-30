package testutils

import (
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type DirContent interface {
	Copy(path string) error
}

// TempDir is the representation of a temporary directory in
// the OS filesystem, which has a Cleanup method used to
// remove the directory, again, after a test.
type TempDir interface {
	Path() string
	Cleanup() error
}

type tempDir struct {
	path string
}

// NewTempDir creates a new temporary directory with some
// initial content.
func NewTempDir(content ...DirContent) (TempDir, error) {
	path, err := os.MkdirTemp("", "temp-*")
	if err != nil {
		return nil, err
	}
	for _, c := range content {
		err := c.Copy(path)
		if err != nil {
			os.RemoveAll(path)
			return nil, err
		}
	}
	return &tempDir{path}, nil
}

func (t *tempDir) Path() string {
	return t.path
}

func (t *tempDir) Cleanup() error {
	if t.path == "" {
		return nil
	}
	defer func() {
		t.path = ""
	}()
	return os.RemoveAll(t.path)
}

////////////////////////////////////////////////////////////////////////////////

type dirContent struct {
	src string
	dst string
}

// WithDirContent populates the temporary directory with some
// content provided by a directory.
// Optionally a target path can be given. The target path MUST be a
// path in the temporary directory. If an absolute path is given
// it is used relative to the temporary directory.
func WithDirContent(src string, dst ...string) DirContent {
	return &dirContent{
		src, general.Optional(dst...),
	}
}

func (d *dirContent) Copy(path string) error {
	dst, err := targetPath(path, d.dst)
	if err != nil {
		return err
	}
	return vfs.CopyDir(osfs.OsFs, d.src, osfs.OsFs, dst)
}

////////////////////////////////////////////////////////////////////////////////

type fileContent dirContent

// WithFileContent populates the temporary directory with some
// file. The name of the given file is preserved.
// Optionally a target path can be given. The target path MUST be a
// path in the temporary directory. If an absolute path is given
// it is used relative to the temporary directory.
func WithFileContent(src string, dst ...string) DirContent {
	return &fileContent{
		src, general.Optional(dst...),
	}
}

func (d *fileContent) Copy(path string) error {
	var err error

	dst := path
	if d.dst != "" {
		dst, err = targetPath(dst, d.dst)
		if err != nil {
			return err
		}
	} else {
		dst = filepath.Join(dst, filepath.Base(d.src))
	}
	err = os.MkdirAll(filepath.Dir(dst), 0o700)
	if err != nil {
		return err
	}
	return vfs.CopyFile(osfs.OsFs, d.src, osfs.OsFs, dst)
}

func targetPath(dst, sub string) (string, error) {
	if sub == "" {
		return dst, nil
	}
	c := filepath.Clean(sub)
	_, l, _ := filepath.SplitPath(c)
	if len(l) > 0 && l[0] == ".." {
		return "", fmt.Errorf("destination above root")
	}
	vol := filepath.VolumeName(sub)
	if len(vol) != 0 {
		return "", fmt.Errorf("destination with volume not possible")
	}
	for strings.HasPrefix(filepath.ToSlash(sub), "/") {
		sub = sub[1:]
	}
	return filepath.Join(dst, sub), nil
}
