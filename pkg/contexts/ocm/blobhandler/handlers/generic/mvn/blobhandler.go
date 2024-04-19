package mvn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/mvn/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/mvn"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
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
		return nil, nil
	}
	if !mvn.IsMimeTypeSupported(mimeType) {
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

	switch mimeType {
	case mime.MIME_TGZ:
		// TODO extract the archive and upload the content
	default:
		err = deploy(artifact, b.spec.Url, blobReader, username, password, blob.Digest())
	}

	// https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact-apis
	// vs. https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifacts-from-archive
	// Headers: X-Explode-Archive: true
	// Map<String, String> headers = ['X-Checksum-Deploy': "true", 'X-Checksum-Sha1': sha1]
	// -H "X-Checksum-Md5: $md5Value" \
	// -H "X-Checksum-Sha1: $sha1Value" \
	// -H "X-Checksum-Sha256: $sha256Value" \

	log.Debug("successfully uploaded")
	return mvn.New(b.spec.Url, artifact.GroupId, artifact.ArtifactId, artifact.Version, mvn.WithClassifier(artifact.Classifier), mvn.WithExtension(artifact.Extension)), nil
}

func ChecksumHeader(digest digest.Digest) string {
	a := digest.Algorithm().String()
	return "X-Checksum-" + strings.ToUpper(a[:1]) + a[1:]
}

func deploy(artifact *mvn.Artifact, url string, reader io.ReadCloser, username string, password string, digest digest.Digest) error {
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
