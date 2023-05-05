// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
)

// ExtractTarToFs writes a tar stream to a filesystem.
func ExtractTarToFs(fs vfs.FileSystem, in io.Reader) error {
	tr := tar.NewReader(in)
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := fs.MkdirAll(header.Name, vfs.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("unable to create directory %s: %w", header.Name, err)
			}
		case tar.TypeReg:
			dir := vfs.Dir(fs, header.Name)
			if err := fs.MkdirAll(dir, 0o766); err != nil {
				return fmt.Errorf("unable to create directory %s: %w", dir, err)
			}
			file, err := fs.OpenFile(header.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, vfs.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("unable to open file %s: %w", header.Name, err)
			}
			//nolint:gosec // We don't know what size limit we could set, the tar
			// archive can be an image layer and that can even reach the gigabyte range.
			// For now, we acknowledge the risk.
			//
			// We checked other softwares and tried to figure out how they manage this,
			// but it's handled the same way.
			if _, err := io.Copy(file, tr); err != nil {
				return fmt.Errorf("unable to copy tar file to filesystem: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("unable to close file %s: %w", header.Name, err)
			}
		}
	}
}

func CreateTarFromFs(fs vfs.FileSystem, path string, fss ...vfs.FileSystem) (err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	tfs := OptionalDefaulted(osfs.New(), fss...)

	w, err := tfs.OpenFile(path, vfs.O_CREATE|vfs.O_TRUNC|vfs.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	finalize.Close(w)
	cw := gzip.NewWriter(w)
	finalize.Close(cw)
	return PackFsIntoTar(fs, cw)
}

func PackFsIntoTar(fs vfs.FileSystem, w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	return vfs.Walk(fs, "", func(path string, info os.FileInfo, err error) error {
		if path == "" {
			return nil
		}
		header := &tar.Header{
			Name:    path,
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}
		if info.IsDir() {
			header.Typeflag = tar.TypeDir
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
		} else {
			fi, err := fs.Lstat(path)
			if err != nil {
				return errors.Wrapf(err, "cannot stat file %q", path)
			}
			if fi.Mode()&os.ModeType == 0 {
				header.Typeflag = tar.TypeReg
				header.Size = fi.Size()
				if err := tw.WriteHeader(header); err != nil {
					return err
				}
				r, err := fs.Open(path)
				if err != nil {
					return errors.Wrapf(err, "cannot open file %q", path)
				}
				if _, err := io.Copy(tw, r); err != nil {
					r.Close()
					return errors.Wrapf(err, "unable to write file: %s", path)
				}
				r.Close()
			}
		}
		return nil
	})
}
