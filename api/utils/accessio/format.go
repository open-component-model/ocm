package accessio

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils/compression"
	"ocm.software/ocm/api/utils/tarutils"
)

const KIND_FILEFORMAT = "file format"

type FileFormat string

func (f FileFormat) String() string {
	return string(f)
}

func (f FileFormat) Suffix() string {
	return suffixes[f]
}

func (o FileFormat) ApplyOption(options Options) error {
	if o != "" {
		options.SetFileFormat(o)
	}
	return nil
}

const (
	FormatTar       FileFormat = "tar"
	FormatTGZ       FileFormat = "tgz"
	FormatDirectory FileFormat = "directory"
	FormatNone      FileFormat = ""
)

var suffixes = map[FileFormat]string{
	FormatTar: "." + string(FormatTar),
	FormatTGZ: "." + string(FormatTGZ),
}

func ErrInvalidFileFormat(fmt string) error {
	return errors.ErrInvalid(KIND_FILEFORMAT, fmt)
}

////////////////////////////////////////////////////////////////////////////////

func GetFormats() []string {
	return []string{string(FormatDirectory), string(FormatTar), string(FormatTGZ)}
}

func GetFormatsFor[T any](fileFormats map[FileFormat]T) []string {
	var def FileFormat

	list := []string{}
	for k := range fileFormats {
		// as favorite default, directory should be the first entry in the list
		if k != FormatDirectory {
			list = append(list, string(k))
		} else {
			def = k
		}
	}
	sort.Strings(list)
	if def != "" {
		return append(append(list[:0:0], string(def)), list...)
	}
	return list
}

// FileFormatForTypeSpec returns the format hint provided
// by a type specification.The format hint is an optional
// suffix separated by a +.
func FileFormatForTypeSpec(t string) FileFormat {
	i := strings.Index(t, "+")
	if i < 0 {
		return ""
	}
	return FileFormat(t[i+1:])
}

// TypeForTypeSpec returns the pure type info provided
// by a type specification.The format hint is an optional
// suffix separated by a +.
func TypeForTypeSpec(t string) string {
	i := strings.Index(t, "+")
	if i < 0 {
		return t
	}
	return t[:i]
}

////////////////////////////////////////////////////////////////////////////////

func CopyFileSystem(format FileFormat, srcfs vfs.FileSystem, src string, dstfs vfs.FileSystem, dst string, perm vfs.FileMode) error {
	compr := compression.None
	switch format {
	case FormatDirectory:
		return vfs.CopyDir(srcfs, src, dstfs, dst)
	case FormatTGZ:
		compr = compression.Gzip
		fallthrough
	case FormatTar:
		file, err := dstfs.OpenFile(dst, vfs.O_CREATE|vfs.O_TRUNC|vfs.O_WRONLY, perm)
		if err != nil {
			return err
		}
		defer file.Close()
		w, err := compr.Compressor(file, nil, nil)
		if err != nil {
			return err
		}
		return tarutils.PackFsIntoTar(srcfs, src, w, tarutils.TarFileSystemOptions{})
	default:
		return errors.ErrUnknown(KIND_FILEFORMAT, format.String())
	}
}

////////////////////////////////////////////////////////////////////////////////

func DetectFormat(path string, fs vfs.FileSystem) (*FileFormat, error) {
	if fs == nil {
		fs = _osfs
	}

	fi, err := fs.Stat(path)
	if err != nil {
		return nil, err
	}

	format := FormatDirectory
	if !fi.IsDir() {
		file, err := fs.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return DetectFormatForFile(file)
	}
	return &format, nil
}

func DetectFormatForFile(file vfs.File) (*FileFormat, error) {
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	format := FormatDirectory
	if !fi.IsDir() {
		var r io.Reader

		defer file.Seek(0, io.SeekStart)
		zip, err := gzip.NewReader(file)
		if err == nil {
			format = FormatTGZ
			defer zip.Close()
			r = zip
		} else {
			file.Seek(0, io.SeekStart)
			format = FormatTar
			r = file
		}
		t := tar.NewReader(r)
		_, err = t.Next()
		if err != nil {
			return nil, err
		}
	}
	return &format, nil
}
