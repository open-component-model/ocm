package maven

import (
	"crypto"

	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/ioutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	mavenblob "github.com/open-component-model/ocm/pkg/blobaccess/maven"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/maven/identity"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/maven"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const BlobHandlerName = "ocm/" + resourcetypes.MAVEN_PACKAGE

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{repospec}
}

var log = logging.DynamicLogger(identity.REALM)

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, resourceType string, hint string, _ cpi.AccessSpec, ctx cpi.StorageContext) (_ cpi.AccessSpec, rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	if hint == "" {
		log.Warn("maven package hint is empty, skipping upload")
		return nil, nil
	}
	// check conditions
	if b.spec == nil {
		return nil, nil
	}
	mimeType := blob.MimeType()
	if resourcetypes.MAVEN_PACKAGE != resourceType {
		log.Debug("not a MVN artifact", "resourceType", resourceType)
		return nil, nil
	}
	if mime.MIME_TGZ != mimeType {
		log.Debug("not a tarball, can't be a complete maven GAV", "mimeType", mimeType)
		return nil, nil
	}

	repo, err := b.spec.GetRepository(ctx.GetContext())
	if err != nil {
		return nil, err
	}

	// setup logger
	log := log.WithValues("repository", repo.String())
	// identify artifact
	coords, err := maven.Parse(hint)
	if err != nil {
		return nil, err
	}
	if !coords.IsPackage() {
		return nil, nil
	}
	log = log.WithValues("groupId", coords.GroupId, "artifactId", coords.ArtifactId, "version", coords.Version)
	log.Debug("identified")

	blobReader, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	finalize.Close(blobReader)
	tempFs, err := tarutils.ExtractTgzToTempFs(blobReader)
	if err != nil {
		return nil, err
	}
	finalize.With(func() error { return vfs.Cleanup(tempFs) })
	files, err := tarutils.ListSortedFilesInDir(tempFs, "", false)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		loop := finalize.Nested()
		log.Debug("uploading", "file", file)
		err := coords.SetClassifierExtensionBy(file)
		if err != nil {
			return nil, err
		}
		readHash, err := tempFs.Open(file)
		if err != nil {
			return nil, err
		}
		loop.Close(readHash)
		// MD5 + SHA1 are still the most used ones in the maven context
		hr := iotools.NewHashReader(readHash, crypto.SHA256, crypto.SHA1, crypto.MD5)
		_, err = hr.CalcHashes()
		if err != nil {
			return nil, err
		}
		reader, err := ioutils.NewDupReadCloser(tempFs.Open(file))
		if err != nil {
			return nil, err
		}
		loop.Close(reader)
		creds, err := mavenblob.GetCredentials(ctx.GetContext(), repo, coords.GroupId)
		if err != nil {
			return nil, err
		}
		err = repo.Upload(coords, reader, creds, hr.Hashes())
		if err != nil {
			return nil, err
		}
		err = loop.Finalize()
		if err != nil {
			return nil, err
		}
	}

	log.Debug("done", "artifact", coords)
	url, err := repo.Url()
	if err != nil {
		return nil, err
	}
	return access.New(url, coords.GroupId, coords.ArtifactId, coords.Version), nil
}
