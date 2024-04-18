package mvn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, resourceType string, hint string, _ cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	// check conditions
	if b.spec == nil {
		return nil, nil
	}
	mimeType := blob.MimeType()
	if resourcetypes.MVN_ARTIFACT != resourceType {
		return nil, nil
	}
	if mime.MIME_JAR != mimeType && mime.MIME_TGZ != mimeType {
		return nil, nil
	}
	if b.spec.Url == "" {
		return nil, fmt.Errorf("MVN repository url not provided")
	}

	// setup logger
	log := logging.Context().Logger(identity.REALM)
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
	case mime.MIME_JAR:
	case mime.MIME_TGZ:

	}

	// https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact-apis
	// vs. https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifacts-from-archive
	// Headers: X-Explode-Archive: true
	// Map<String, String> headers = ['X-Checksum-Deploy': "true", 'X-Checksum-Sha1': sha1]
	// -H "X-Checksum-Md5: $md5Value" \
	// -H "X-Checksum-Sha1: $sha1Value" \
	// -H "X-Checksum-Sha256: $sha256Value" \
	// Headers: X-Checksum-Deploy: true, X-Checksum-Sha1: sha1Value, X-Checksum-Sha256: sha256Value, X-Checksum: checksum value (type is resolved by length)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, b.spec.Url+"/"+artifact.Path(), blobReader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	//req.Header.Set("X-Checksum", )

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusCreated {
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("http (%d) - failed to upload artifact: %s", resp.StatusCode, string(all))
	}

	// Validate the response - especially the hash values with the ones we've tried to send
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var artifactBody Body
	err = json.Unmarshal(respBody, &artifactBody)
	if err != nil {
		return nil, err
	}
	blobDigest := blob.Digest()
	remoteDigest := artifactBody.Checksums[string(blobDigest.Algorithm())]
	if remoteDigest == "" {
		log.Warn("no checksum found for algorithm, we can't guarantee that the artifact has been uploaded correctly", "algorithm", blobDigest.Algorithm())
	} else if remoteDigest != blobDigest.Encoded() {
		return nil, fmt.Errorf("failed to upload artifact: checksums do not match")
	}

	log.Debug("successfully uploaded")
	return mvn.New(b.spec.Url, artifact.GroupId, artifact.ArtifactId, artifact.Version, mvn.WithClassifier(artifact.Classifier), mvn.WithExtension(artifact.Extension)), nil
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
