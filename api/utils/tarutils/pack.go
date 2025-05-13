package tarutils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	pathutil "path"
	"sort"
	"strings"
	"time"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func CreateTarFromFs(fs vfs.FileSystem, path string, compress func(w io.Writer) io.WriteCloser, fss ...vfs.FileSystem) (err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	tfs := general.OptionalDefaulted(osfs.New(), fss...)

	f, err := tfs.OpenFile(path, vfs.O_CREATE|vfs.O_TRUNC|vfs.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	finalize.Close(f)
	var w io.Writer
	if compress != nil {
		cw := compress(f)
		finalize.Close(cw)
		w = cw
	} else {
		w = f
	}
	return PackFsIntoTar(fs, "", w, TarFileSystemOptions{})
}

// TarFileSystemOptions describes additional options for tarring a filesystem.
type TarFileSystemOptions struct {
	IncludeFiles []string
	ExcludeFiles []string
	// PreserveDir defines that the directory specified in the Path field should be included in the blob.
	// Only supported for Type dir.
	PreserveDir    bool
	FollowSymlinks bool

	root string
}

// Included determines whether a file should be included.
func (opts *TarFileSystemOptions) Included(path string) (bool, error) {
	// if a root path is given remove it rom the path to be checked
	if len(opts.root) != 0 {
		path = strings.TrimPrefix(path, opts.root)
	}
	// first check if a exclude regex matches
	for _, ex := range opts.ExcludeFiles {
		match, err := filepath.Match(ex, path)
		if err != nil {
			return false, fmt.Errorf("malformed filepath syntax %q", ex)
		}
		if match {
			return false, nil
		}
	}

	// if no includes are defined, include all files
	if len(opts.IncludeFiles) == 0 {
		return true, nil
	}
	// otherwise check if the file should be included
	for _, in := range opts.IncludeFiles {
		match, err := filepath.Match(in, path)
		if err != nil {
			return false, fmt.Errorf("malformed filepath syntax %q", in)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

// PackFsIntoTar creates a tar archive from a filesystem.
func PackFsIntoTar(fs vfs.FileSystem, root string, writer io.Writer, opts TarFileSystemOptions) error {
	tw := tar.NewWriter(writer)
	if opts.PreserveDir {
		opts.root = pathutil.Base(root)
	}
	if err := addFileToTar(fs, tw, opts.root, root, opts); err != nil {
		return err
	}
	return tw.Close()
}

func addFileToTar(fs vfs.FileSystem, tw *tar.Writer, path string, realPath string, opts TarFileSystemOptions) error {
	if len(path) != 0 { // do not check the root
		include, err := opts.Included(path)
		if err != nil {
			return err
		}
		if !include {
			return nil
		}
	}
	info, err := fs.Lstat(realPath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path

	// TODO: @jakobmoellerdev pass options to omit modtime

	switch {
	case info.IsDir():
		// do not write root header
		if len(path) != 0 {
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("unable to write header for %q: %w", path, err)
			}
		}
		err := vfs.Walk(fs, realPath, func(subFilePath string, info vfs.FileInfo, err error) error {
			if subFilePath == realPath {
				return nil
			}
			if err != nil {
				return err
			}
			/*
				relPath, err := vfs.Rel(fs, realPath, subFilePath)
				if err != nil {
					return fmt.Errorf("unable to calculate relative path for %s: %w", subFilePath, err)
				}*/
			relPath := vfs.Base(fs, subFilePath)
			err = addFileToTar(fs, tw, vfs.Join(fs, path, relPath), subFilePath, opts)
			if err != nil {
				return fmt.Errorf("failed to tar the input from %q: %w", subFilePath, err)
			}
			if info.IsDir() {
				return vfs.SkipDir
			}
			return nil
		})
		return err
	case info.Mode().IsRegular():
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write header for %q: %w", path, err)
		}
		file, err := fs.OpenFile(realPath, vfs.O_RDONLY, vfs.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open file %q: %w", path, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			_ = file.Close()
			return fmt.Errorf("unable to add file to tar %q: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("unable to close file %q: %w", path, err)
		}
		return nil
	case header.Typeflag == tar.TypeSymlink:
		if !opts.FollowSymlinks {
			link, err := fs.Readlink(realPath)
			if err != nil {
				return fmt.Errorf("cannot read symlink %s: %w", realPath, err)
			}
			header.Linkname = link
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("unable to write header for %q: %w", path, err)
			}
			return nil
		} else {
			effPath, err := vfs.EvalSymlinks(fs, realPath)
			if err != nil {
				return fmt.Errorf("unable to follow symlink %s: %w", realPath, err)
			}
			return addFileToTar(fs, tw, path, effPath, opts)
		}
	default:
		return fmt.Errorf("unsupported file type %s in %s", info.Mode().String(), path)
	}
}

func Epoch() time.Time {
	return time.Unix(0, 0)
}

func SimpleTarHeader(fs vfs.FileSystem, filepath string) (*tar.Header, error) {
	info, err := fs.Lstat(filepath)
	if err != nil {
		return nil, err
	}
	return RegularFileInfoHeader(info), nil
}

// RegularFileInfoHeader creates a tar header for a regular file (`tar.TypeReg`).
// Besides name and size, the other header fields are set to default values (`fs.ModePerm`, 0, "", `time.Unix(0,0)`).
func RegularFileInfoHeader(fi fs.FileInfo) *tar.Header {
	h := &tar.Header{
		Typeflag:   tar.TypeReg,
		Name:       fi.Name(),
		Size:       fi.Size(),
		Mode:       int64(fs.ModePerm),
		Uid:        0,
		Gid:        0,
		Uname:      "",
		Gname:      "",
		ModTime:    Epoch(),
		AccessTime: Epoch(),
		ChangeTime: Epoch(),
	}
	return h
}

// ListSortedFilesInDir returns a list of files in a directory sorted by name.
// Attention: If 'flat == true', files with same name but in different sub-paths, will be listed only once!!!
func ListSortedFilesInDir(fs vfs.FileSystem, root string, flat bool) ([]string, error) {
	var files []string
	err := vfs.Walk(fs, root, func(path string, info vfs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			pathOrName := path
			if flat {
				pathOrName = info.Name()
			}
			files = append(files, pathOrName)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// TgzFs creates a tar.gz archive from a filesystem with all files being in the root of the zipped archive.
// The writer is closed after the archive is written. The TAR-headers are normalized, see RegularFileInfoHeader.
func TgzFs(fs vfs.FileSystem, writer io.Writer) error {
	zip := gzip.NewWriter(writer)
	err := TarFlatFs(fs, zip)
	if err != nil {
		return err
	}
	return zip.Close()
}

func TarFlatFs(fs vfs.FileSystem, writer io.Writer) error {
	tw := tar.NewWriter(writer)
	files, err := ListSortedFilesInDir(fs, "", true)
	if err != nil {
		return err
	}

	for _, fileName := range files {
		header, err := SimpleTarHeader(fs, fileName)
		if err != nil {
			return err
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write header for %q: %w", fileName, err)
		}
		file, err := fs.OpenFile(fileName, vfs.O_RDONLY, vfs.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open file %q: %w", fileName, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			_ = file.Close()
			return fmt.Errorf("unable to add file to tar %q: %w", fileName, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("unable to close file %q: %w", fileName, err)
		}
	}

	return tw.Close()
}
