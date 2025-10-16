package dirtree

import (
	"compress/gzip"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
)

func DataAccess(path string, opts ...Option) (bpi.DataAccess, error) {
	blobAccess, err := BlobAccess(path, opts...)
	if err != nil {
		return nil, err
	}
	return blobAccess, nil
}

func BlobAccess(path string, opts ...Option) (_ bpi.BlobAccess, rerr error) {
	eff := optionutils.EvalOptions(opts...)
	fs := utils.FileSystem(eff.FileSystem)

	ok, err := vfs.IsDir(fs, path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%q is no directory", path)
	}

	taropts := tarutils.TarFileSystemOptions{
		IncludeFiles:   eff.IncludeFiles,
		ExcludeFiles:   eff.ExcludeFiles,
		PreserveDir:    utils.AsBool(eff.PreserveDir),
		FollowSymlinks: utils.AsBool(eff.FollowSymlinks),
	}

	temp, err := file.NewTempFile(fs.FSTempDir(), "resourceblob*.tgz", fs)
	if err != nil {
		return nil, err
	}
	defer errors.PropagateError(&rerr, temp.Close)

	if utils.AsBool(eff.CompressWithGzip) {
		if eff.MimeType == "" {
			eff.MimeType = mime.MIME_TGZ
		}
		gw := gzip.NewWriter(temp.Writer())
		if err := tarutils.PackFsIntoTar(fs, path, gw, taropts); err != nil {
			return nil, fmt.Errorf("unable to tar input artifact: %w", err)
		}
		if err := gw.Close(); err != nil {
			return nil, fmt.Errorf("unable to close gzip writer: %w", err)
		}
	} else {
		if eff.MimeType == "" {
			eff.MimeType = mime.MIME_TAR
		}
		if err := tarutils.PackFsIntoTar(fs, path, temp.Writer(), taropts); err != nil {
			return nil, fmt.Errorf("unable to tar input artifact: %w", err)
		}
	}
	return temp.AsBlob(eff.MimeType), nil
}

func Provider(path string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		return BlobAccess(path, opts...)
	})
}
