package dirtree

import (
	"compress/gzip"
	"fmt"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
	"os"
)

func DataAccessForDirTree(path string, opts ...Option) (accessio.DataAccess, error) {
	blobAccess, err := BlobAccessForDirTree(mime.MIME_TAR, path, opts...)
	if err != nil {
		return nil, err
	}
	return blobAccess, nil
}

func BlobAccessForDirTree(mimeType string, path string, opts ...Option) (_ accessio.BlobAccess, rerr error) {
	var eff Options
	for _, opt := range opts {
		opt.ApplyToDirtreeOptions(&eff)
	}
	taropts := tarutils.TarFileSystemOptions{
		IncludeFiles:   eff.IncludeFiles,
		ExcludeFiles:   eff.ExcludeFiles,
		PreserveDir:    utils.AsBool(eff.PreserveDir),
		FollowSymlinks: utils.AsBool(eff.FollowSymlinks),
	}

	var fs vfs.FileSystem
	var tempdir string
	if eff.FileSystem != nil {
		fs = eff.FileSystem
		tempdir = ""
	} else {
		fs = osfs.New()
		tempdir = os.TempDir()
	}

	temp, err := accessio.NewTempFile(fs, tempdir, "resourceblob*.tgz")
	if err != nil {
		return nil, err
	}
	defer errors.PropagateError(&rerr, temp.Close)

	if utils.AsBool(eff.CompressWithGzip) {
		if mimeType == "" {
			mimeType = mime.MIME_TGZ
		}
		gw := gzip.NewWriter(temp.Writer())
		if err := tarutils.PackFsIntoTar(fs, path, gw, taropts); err != nil {
			return nil, fmt.Errorf("unable to tar input artifact: %w", err)
		}
		if err := gw.Close(); err != nil {
			return nil, fmt.Errorf("unable to close gzip writer: %w", err)
		}
	} else {
		if mimeType == "" {
			mimeType = mime.MIME_TAR
		}
		if err := tarutils.PackFsIntoTar(fs, path, temp.Writer(), taropts); err != nil {
			return nil, fmt.Errorf("unable to tar input artifact: %w", err)
		}
	}
	return temp.AsBlob(mimeType), nil
}
