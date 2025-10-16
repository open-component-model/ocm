package maven

import (
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/tech/maven/identity"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
)

type coords = *maven.Coordinates

type spec struct {
	coords
	repo    *maven.Repository
	options *Options
}

func (s *spec) getBlobAccess() (_ bpi.BlobAccess, rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	log := s.options.Logger("RepoUrl", s.repo.String())
	creds, err := s.options.GetCredentials(s.repo, s.GroupId)
	if err != nil {
		return nil, err
	}
	fileMap, err := s.repo.GavFiles(s.coords, creds)
	if err != nil {
		return nil, err
	}

	fileMap = s.coords.FilterFileMap(fileMap)

	switch l := len(fileMap); {
	case l <= 0:
		return nil, errors.New("no maven artifact files found")
	case l == 1 && optionutils.AsValue(s.Extension) != "" && s.Classifier != nil:
		for file, hash := range fileMap {
			metadata, err := s.repo.GetFileMeta(s.coords, file, hash, creds)
			if err != nil {
				return nil, err
			}
			return blobAccessForRepositoryAccess(metadata, creds, s.options)
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
		metadata, err := s.repo.GetFileMeta(s.coords, file, hash, creds)
		if err != nil {
			return nil, err
		}

		// download the artifact into the temporary file system
		out, err := tmpfs.Create(file)
		if err != nil {
			return nil, err
		}
		loop.Close(out)

		reader, err := metadata.Location.GetReader(creds)
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
	fs := s.options.GetCachingFileSystem()
	tgz, err := vfs.TempFile(fs, "", "maven-"+s.coords.FileNamePrefix()+"-*.tar.gz")
	if err != nil {
		return nil, err
	}

	dw := iotools.NewDigestWriterWith(digest.SHA256, tgz)
	finalize.Close(dw)

	err = tarutils.TgzFlatFs(tmpfs, dw)
	if err != nil {
		return nil, err
	}
	log.Debug("created", "file", tgz.Name())
	return file.BlobAccessForTemporaryFilePath(mime.MIME_TGZ, tgz.Name(), file.WithFileSystem(fs), file.WithDigest(dw.Digest()), file.WithSize(dw.Size())), nil
}

func blobAccessForRepositoryAccess(meta *BlobMeta, creds maven.Credentials, opts *Options) (bpi.BlobAccess, error) {
	reader := func() (io.ReadCloser, error) {
		return meta.Location.GetReader(creds)
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
	acc := blobaccess.DataAccessForReaderFunction(reader, meta.Location.String())
	return accessobj.CachedBlobAccessForWriterWithCache(opts.Cache(), meta.MimeType, accessio.NewDataAccessWriter(acc)), nil
}

func MapCredentials(creds credentials.Credentials) maven.Credentials {
	if creds == nil || (!creds.ExistsProperty(identity.ATTR_USERNAME) && !creds.ExistsProperty(identity.ATTR_PASSWORD)) {
		return nil
	}
	return &maven.BasicAuthCredentials{
		Username: creds.GetProperty(identity.ATTR_USERNAME),
		Password: creds.GetProperty(identity.ATTR_PASSWORD),
	}
}

func GetCredentials(ctx credentials.ContextProvider, repo *Repository, groupId string) (maven.Credentials, error) {
	consumerid, err := identity.GetConsumerId(repo.String(), groupId)
	if err != nil {
		return nil, err
	}
	creds, err := credentials.CredentialsForConsumer(ctx, consumerid, identity.IdentityMatcher)
	if err != nil {
		return nil, err
	}
	return MapCredentials(creds), nil
}
