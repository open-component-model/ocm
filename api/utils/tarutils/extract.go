package tarutils

import (
	"archive/tar"
	"fmt"
	"io"
	"math"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/compression"
)

// ExtractArchiveToFs wunpacks an archive to a filesystem.
func ExtractArchiveToFs(fs vfs.FileSystem, path string, fss ...vfs.FileSystem) error {
	sfs := utils.OptionalDefaulted(osfs.New(), fss...)

	f, err := sfs.Open(path)
	if err != nil {
		return errors.Wrapf(err, "cannot open %s", path)
	}
	defer f.Close()
	r, _, err := compression.AutoDecompress(f)
	if err != nil {
		return errors.Wrapf(err, "cannot determine compression for %s", path)
	}
	return ExtractTarToFs(fs, r)
}

// ExtractArchiveToFsWithInfo unpacks an archive to a filesystem.
func ExtractArchiveToFsWithInfo(fs vfs.FileSystem, path string, fss ...vfs.FileSystem) (int64, int64, error) {
	sfs := utils.OptionalDefaulted(osfs.New(), fss...)

	f, err := sfs.Open(path)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "cannot open %s", path)
	}
	defer f.Close()
	r, _, err := compression.AutoDecompress(f)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "cannot determine compression for %s", path)
	}
	return ExtractTarToFsWithInfo(fs, r)
}

// ExtractTarToFs writes a tar stream to a filesystem.
func ExtractTarToFs(fs vfs.FileSystem, in io.Reader) error {
	_, _, err := ExtractTarToFsWithInfo(fs, in)
	return err
}

// UnzipTarToFs tries to decompress the input stream and then writes the tar stream to a filesystem.
func UnzipTarToFs(fs vfs.FileSystem, in io.Reader) error {
	r, _, err := compression.AutoDecompress(in)
	if err != nil {
		return err
	}
	defer r.Close()
	err = ExtractTarToFs(fs, r)
	if err != nil {
		return err
	}
	return err
}

// ExtractTgzToTempFs extracts a tar.gz archive to a temporary filesystem.
// You should call vfs.Cleanup on the returned filesystem to clean up the temporary files.
func ExtractTgzToTempFs(in io.Reader) (vfs.FileSystem, error) {
	fs, err := osfs.NewTempFileSystem()
	if err != nil {
		return nil, err
	}

	return fs, UnzipTarToFs(fs, in)
}

func ExtractTarToFsWithInfo(fs vfs.FileSystem, in io.Reader) (fcnt int64, bcnt int64, err error) {
	tr := tar.NewReader(in)
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fcnt, bcnt, nil
			}
			return fcnt, bcnt, err
		}

		if header.Mode < 0 || header.Mode > math.MaxUint32 {
			return fcnt, bcnt, fmt.Errorf("file %s: mode %d out of range for uint32", header.Name, header.Mode)
		}
		fileMode := vfs.FileMode(header.Mode)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := fs.MkdirAll(header.Name, fileMode); err != nil {
				return fcnt, bcnt, fmt.Errorf("unable to create directory %s: %w", header.Name, err)
			}
		case tar.TypeSymlink, tar.TypeLink:
			dir := vfs.Dir(fs, header.Name)
			if err := fs.MkdirAll(dir, 0o766); err != nil {
				return fcnt, bcnt, fmt.Errorf("unable to create directory %s: %w", dir, err)
			}
			err := fs.Symlink(header.Linkname, header.Name)
			if err != nil {
				return fcnt, bcnt, fmt.Errorf("unable to create symbolic link %s: %w", header.Name, err)
			}
			fcnt++
		case tar.TypeReg:
			dir := vfs.Dir(fs, header.Name)
			if err := fs.MkdirAll(dir, 0o766); err != nil {
				return fcnt, bcnt, fmt.Errorf("unable to create directory %s: %w", dir, err)
			}
			file, err := fs.OpenFile(header.Name, vfs.O_WRONLY|vfs.O_CREATE|vfs.O_TRUNC, fileMode)
			if err != nil {
				return fcnt, bcnt, fmt.Errorf("unable to open file %s: %w", header.Name, err)
			}
			bcnt += header.Size
			//nolint:gosec // We don't know what size limit we could set, the tar
			// archive can be an image layer and that can even reach the gigabyte range.
			// For now, we acknowledge the risk.
			//
			// We checked other software and tried to figure out how they manage this,
			// but it's handled the same way.
			if _, err := io.Copy(file, tr); err != nil {
				return fcnt, bcnt, fmt.Errorf("unable to copy tar file to filesystem: %w", err)
			}
			if err := file.Close(); err != nil {
				return fcnt, bcnt, fmt.Errorf("unable to close file %s: %w", header.Name, err)
			}
			fcnt++
		}
	}
}
