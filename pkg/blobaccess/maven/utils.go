// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/optionutils"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
	"github.com/opencontainers/go-digest"
	"io"
)

type coords = *maven.Coordinates

type spec struct {
	coords
	repoUrl string
	options *Options
}

func (s *spec) getBlobAccess() (_ bpi.BlobAccess, rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	log := s.options.Logger("BaseUrl", s.repoUrl)
	creds, err := s.options.GetCredentials(s.repoUrl, s.GroupId)
	if err != nil {
		return nil, err
	}
	fileMap, err := maven.GavFiles(s.repoUrl, s.coords, creds, s.options.FileSystem)
	if err != nil {
		return nil, err
	}

	fileMap = s.coords.FilterFileMap(fileMap)

	switch l := len(fileMap); {
	case l <= 0:
		return nil, errors.New("no maven artifact files found")
	case l == 1 && optionutils.AsValue(s.Extension) != "" && s.Classifier != nil:
		for file, hash := range fileMap {
			metadata, err := maven.GetFileMeta(s.repoUrl, s.coords, file, hash, creds, s.options.FileSystem)
			if err != nil {
				return nil, err
			}
			return blobAccessForRepositoryAccess(metadata, creds, s.options, s.options.FileSystem)
		}
		// default: continue below with: create tmpfs where all files can be downloaded to and packed together as tar.gz
	}

	tmpfs, err := osfs.NewTempFileSystem()
	if err != nil {
		return nil, err
	}
	finalize.With(func() error {
		return vfs.Cleanup(tmpfs)
	})

	for file, hash := range fileMap {
		loop := finalize.Nested()
		metadata, err := maven.GetFileMeta(s.repoUrl, s.coords, file, hash, creds, s.options.FileSystem)
		if err != nil {
			return nil, err
		}

		// download the artifact into the temporary file system
		out, err := tmpfs.Create(file)
		if err != nil {
			return nil, err
		}
		loop.Close(out)

		reader, err := maven.GetReader(metadata.Url, creds, s.options.FileSystem)
		if err != nil {
			return nil, err
		}
		loop.Close(reader)
		if hash > 0 {
			dreader := iotools.NewDigestReaderWithHash(hash, reader)
			_, err = io.Copy(out, dreader)
			if err != nil {
				return nil, err
			}
			sum := dreader.Digest().Encoded()
			if metadata.Hash != sum {
				return nil, errors.Newf("%s digest mismatch: expected %s, found %s", metadata.HashType, metadata.Hash, sum)
			}
		} else {
			_, err = io.Copy(out, reader)
			return nil, err
		}
		err = loop.Finalize()
		if err != nil {
			return nil, err
		}
	}

	// pack all downloaded files into a tar.gz file
	fs := utils.FileSystem(s.options.FileSystem)
	tgz, err := vfs.TempFile(fs, "", "maven-"+s.coords.FileNamePrefix()+"-*.tar.gz")
	if err != nil {
		return nil, err
	}

	dw := iotools.NewDigestWriterWith(digest.SHA256, tgz)
	finalize.Close(dw)

	err = tarutils.TgzFs(tmpfs, dw)
	if err != nil {
		return nil, err
	}
	log.Debug("created", "file", tgz.Name())
	return blobaccess.ForTemporaryFilePathWithMeta(mime.MIME_TGZ, dw.Digest(), dw.Size(), tgz.Name(), fs), nil
}

func blobAccessForRepositoryAccess(meta *BlobMeta, creds maven.Credentials, opts *Options, fss ...vfs.FileSystem) (bpi.BlobAccess, error) {
	reader := func() (io.ReadCloser, error) {
		return maven.GetReader(meta.Url, creds, fss...)
	}
	if meta.Hash != "" {
		getreader := reader
		reader = func() (io.ReadCloser, error) {
			readCloser, err := getreader()
			if err != nil {
				return nil, err
			}
			return iotools.VerifyingReaderWithHash(readCloser, meta.HashType, meta.Hash), nil
		}
	}
	acc := blobaccess.DataAccessForReaderFunction(reader, meta.Url)
	return accessobj.CachedBlobAccessForWriterWithCache(opts.Cache(), meta.MimeType, accessio.NewDataAccessWriter(acc)), nil
}
