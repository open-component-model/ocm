package maven

import (
	"crypto"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/ioutils"
	mlog "github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/cpi"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/maven"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/tech/maven/identity"
	mavenblob "ocm.software/ocm/api/utils/blobaccess/maven"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
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

	if err := verifyGavInformation(tempFs, coords, files, log); err != nil {
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

// PomGav defines gav information in a POM file.
type PomGav struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// GAV returns the GAV coordinates of the Maven Coordinates.
func (c *PomGav) GAV() string {
	return c.GroupId + ":" + c.ArtifactId + ":" + c.Version
}

func verifyGavInformation(fs vfs.FileSystem, coords *maven.Coordinates, files []string, log mlog.Logger) error {
	var found string
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file)) == ".pom" {
			found = file
			break
		}
	}

	if found == "" {
		log.Warn("no POM found to verify GAV information")

		return nil
	}

	file, err := fs.Open(found)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read pom file: %w", err)
	}

	pomGav := &PomGav{}
	if err := xml.Unmarshal(content, pomGav); err != nil {
		return fmt.Errorf("failed to marshal pom content: %w", err)
	}

	if pomGav.GAV() != coords.GAV() {
		return fmt.Errorf("%s did not match pom content %s", coords.GAV(), pomGav.GAV())
	}

	return nil
}
