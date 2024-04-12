package mvn

import (
	"encoding/json"
	"fmt"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/mvn/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/mvn"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
	"io"
	"net/http"
)

const BLOB_HANDLER_NAME = "ocm/" + resourcetypes.MVN_ARTIFACT

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{repospec}
}

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, _ string, _ string, _ cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	// check conditions
	if b.spec == nil {
		return nil, nil
	}
	mimeType := blob.MimeType()
	if mime.MIME_JAR != mimeType {
		return nil, nil
	}
	if b.spec.Url == "" {
		return nil, fmt.Errorf("MVN repository url not provided")
	}

	// setup logger
	log := logging.Context().Logger(mvn.REALM)
	log = log.WithValues("repository", b.spec.Url)

	blobReader, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer blobReader.Close()

	/*/ FIXME: do we need to read the whole file into memory? Maybe to guess the groupId, artifactId and version?
	data, err := io.ReadAll(blobReader)
	if err != nil {
		return nil, err
	} */
	var artifact *Artifact
	// FIXME: how do I get the groupId, artifactId and version from accessSpec.AccessMethod().???
	log.Debug("reading jar file - but, where is the groupId, artifactId and version?")
	artifact = &Artifact{
		GroupId:    "ocm.software", // FIXME: hardcoded
		ArtifactId: "hello-ocm",    // FIXME: hardcoded
		Version:    "0.0.1",        // FIXME: hardcoded
		Packaging:  "jar",          // FIXME: hardcoded
	}
	log = log.WithValues("groupId", artifact.GroupId, "artifactId", artifact.ArtifactId, "version", artifact.Version)
	log.Debug("identified")

	// get credentials
	cred := identity.GetCredentials(ctx.GetContext(), b.spec.Url, artifact.GroupPath())
	if cred == nil {
		return nil, fmt.Errorf("no credentials found for %s. Couldn't upload '%s'", b.spec.Url, artifact.GAV())
	}
	log.Debug("found credentials")

	username := cred[identity.ATTR_USERNAME]
	password := cred[identity.ATTR_PASSWORD]
	if username == "" || password == "" {
		return nil, fmt.Errorf("credentials for %s are invalid. Username or password missing! Couldn't upload '%s'", b.spec.Url, artifact.GAV())
	}
	log = log.WithValues("user", username)

	// Create a new request
	req, err := http.NewRequest("PUT", b.spec.Url+"/"+artifact.Path(), blobReader)
	if err != nil {
		panic(err)
	}

	// Add authentication
	req.SetBasicAuth(username, password)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
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

	respBody, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(respBody, &artifact)
	if err != nil {
		return nil, err
	}

	log.Debug("successfully uploaded")
	return mvn.New(b.spec.Url, artifact.GroupId, artifact.ArtifactId, artifact.Version), nil
}
