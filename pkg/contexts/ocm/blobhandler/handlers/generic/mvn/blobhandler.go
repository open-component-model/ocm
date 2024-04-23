package mvn

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/mvn/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/mvn"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const BLOB_HANDLER_NAME = "ocm/" + resourcetypes.MVN_ARTIFACT

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{repospec}
}

var log = logging.Context().Logger(identity.REALM)

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, resourceType string, hint string, _ cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	// check conditions
	if b.spec == nil {
		return nil, nil
	}
	mimeType := blob.MimeType()
	if resourcetypes.MVN_ARTIFACT != resourceType {
		log.Debug("not a MVN artifact", "resourceType", resourceType)
		return nil, nil
	}
	if mime.MIME_TGZ != mimeType {
		log.Debug("not a tarball, can't be a complete mvn GAV", "mimeType", mimeType)
		return nil, nil
	}
	if b.spec.Url == "" {
		return nil, fmt.Errorf("MVN repository url not provided")
	}

	// setup logger
	log = log.WithValues("repository", b.spec.Url)

	// identify artifact
	artifact := mvn.ArtifactFromHint(hint)
	log = log.WithValues("groupId", artifact.GroupId, "artifactId", artifact.ArtifactId, "version", artifact.Version)
	log.Debug("identified")

	// get credentials
	cred := identity.GetCredentials(ctx.GetContext(), b.spec.Url, artifact.GroupPath())
	if cred == nil {
		return nil, fmt.Errorf("no credentials found for %s. Couldn't upload '%s'", b.spec.Url, artifact)
	}
	username := cred[identity.ATTR_USERNAME]
	password := cred[identity.ATTR_PASSWORD]
	if username == "" || password == "" {
		return nil, fmt.Errorf("credentials for %s are invalid. Username or password missing! Couldn't upload '%s'", b.spec.Url, artifact)
	}
	log = log.WithValues("user", username)
	log.Debug("found credentials")

	// Create a new request
	blobReader, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer blobReader.Close()

	tempFs, err := osfs.NewTempFileSystem()
	if err != nil {
		return nil, err
	}
	defer vfs.Cleanup(tempFs)
	err = tarutils.ExtractTarToFs(tempFs, blobReader)
	// err = tarutils.ExtractTarToFs(tempFs, compression.AutoDecompress(blobReader)) // ???
	if err != nil {
		return nil, err
	}
	files, err := tarutils.FlatListSortedFilesInDir(tempFs, "")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		log.Debug("uploading", "file", file)
		artifact = artifact.ClassifierExtensionFrom(file)
		reader, err := tempFs.Open(file)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		hash := sha256.New()
		if _, err := io.Copy(hash, reader); err != nil {
			return nil, err
		}
		err = deploy(artifact, b.spec.Url, reader, username, password, digest.NewDigest(digest.SHA256, hash))
		if err != nil {
			return nil, err
		}
	}
	/*
		default:
			err = deploy(artifact, b.spec.Url, blobReader, username, password, blob.Digest())
			if err != nil {
				return nil, err
			}
		}
	*/

	log.Debug("done", "artifact", artifact)
	return mvn.New(b.spec.Url, artifact.GroupId, artifact.ArtifactId, artifact.Version, mvn.WithClassifier(artifact.Classifier), mvn.WithExtension(artifact.Extension)), nil
}

func ChecksumHeader(digest digest.Digest) string {
	a := digest.Algorithm().String()
	return "X-Checksum-" + strings.ToUpper(a[:1]) + a[1:]
}

func deploy(artifact *mvn.Artifact, url string, reader io.ReadCloser, username string, password string, digest digest.Digest) error {
	// https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact-apis
	// vs. https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifacts-from-archive
	// Headers: X-Checksum-Deploy: true, X-Checksum-Sha1: sha1Value, X-Checksum-Sha256: sha256Value, X-Checksum: checksum value (type is resolved by length)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, artifact.Url(url), reader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("X-Checksum", digest.Encoded())
	req.Header.Set(ChecksumHeader(digest), digest.Encoded())

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusCreated {
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("http (%d) - failed to upload artifact: %s", resp.StatusCode, string(all))
	}
	log.Debug("uploaded", "artifact", artifact, "extension", artifact.Extension, "classifier", artifact.Classifier)

	// Validate the response - especially the hash values with the ones we've tried to send
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var artifactBody Body
	err = json.Unmarshal(respBody, &artifactBody)
	if err != nil {
		return err
	}

	remoteDigest := artifactBody.Checksums[string(digest.Algorithm())]
	if remoteDigest == "" {
		log.Warn("no checksum found for algorithm, we can't guarantee that the artifact has been uploaded correctly", "algorithm", digest.Algorithm())
	} else if remoteDigest != digest.Encoded() {
		return fmt.Errorf("failed to upload artifact: checksums do not match")
	}
	log.Debug("digests are ok", "remoteDigest", remoteDigest, "digest", digest.Encoded())
	return nil
}

// Body is the response struct of a deployment from the MVN repository (JFrog Artifactory).
type Body struct {
	Repo        string            `json:"repo"`
	Path        string            `json:"path"`
	DownloadUri string            `json:"downloadUri"`
	Uri         string            `json:"uri"`
	MimeType    string            `json:"mimeType"`
	Size        string            `json:"size"`
	Checksums   map[string]string `json:"checksums"`
}
