package mvn

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"

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

	tempFs, err := tarutils.ExtractTgzToTempFs(blobReader)
	if err != nil {
		return nil, err
	}
	defer vfs.Cleanup(tempFs)
	files, err := tarutils.ListSortedFilesInDir(tempFs, "", false)
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
		hash, err := GetHash(crypto.SHA256, tempFs, file)
		if err != nil {
			return nil, err
		}
		err = deploy(artifact, b.spec.Url, reader, username, password, crypto.SHA256, hash)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("done", "artifact", artifact)
	return mvn.New(b.spec.Url, artifact.GroupId, artifact.ArtifactId, artifact.Version, mvn.WithClassifier(artifact.Classifier), mvn.WithExtension(artifact.Extension)), nil
}

func GetHash(hash crypto.Hash, fs vfs.FileSystem, path string) (string, error) {
	reader, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	digest := hash.New()
	if _, err := io.Copy(digest, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func ChecksumHeader(hash crypto.Hash) string {
	a := strings.ReplaceAll(hash.String(), "-", "")
	return "X-Checksum-" + a[:1] + strings.ToLower(a[1:])
}

func deploy(artifact *mvn.Artifact, url string, reader io.ReadCloser, username string, password string, hash crypto.Hash, digest string) error {
	// https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact-apis
	// vs. https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifacts-from-archive
	// Headers: X-Checksum-Deploy: true, X-Checksum-Sha1: sha1Value, X-Checksum-Sha256: sha256Value, X-Checksum: checksum value (type is resolved by length)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, artifact.Url(url), reader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("X-Checksum", digest)
	req.Header.Set(ChecksumHeader(hash), digest)

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

	remoteDigest := artifactBody.Checksums[strings.ReplaceAll(strings.ToLower(hash.String()), "-", "")]
	if remoteDigest == "" {
		log.Warn("no checksum found for algorithm, we can't guarantee that the artifact has been uploaded correctly", "algorithm", hash)
	} else if remoteDigest != digest {
		return fmt.Errorf("failed to upload artifact: checksums do not match")
	}
	log.Debug("digests are ok", "remoteDigest", remoteDigest, "digest", digest)
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
