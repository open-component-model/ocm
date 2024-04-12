package npm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	npmCredentials "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/npm/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/npm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
)

const BLOB_HANDLER_NAME = "ocm/" + resourcetypes.NPM_PACKAGE

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{repospec}
}

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, _ string, _ string, _ cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	if b.spec == nil {
		return nil, nil
	}

	mimeType := blob.MimeType()
	if mime.MIME_TGZ != mimeType && mime.MIME_TGZ_ALT != mimeType {
		return nil, nil
	}

	if b.spec.Url == "" {
		return nil, fmt.Errorf("NPM registry url not provided")
	}

	blobReader, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer blobReader.Close()

	data, err := io.ReadAll(blobReader)
	if err != nil {
		return nil, err
	}

	// read package.json from tarball to get name, version, etc.
	log := logging.Context().Logger(npmCredentials.REALM)
	log.Debug("reading package.json from tarball")
	var pkg *Package
	pkg, err = prepare(data)
	if err != nil {
		return nil, err
	}
	tbName := pkg.Name + "-" + pkg.Version + ".tgz"
	pkg.Dist.Tarball = b.spec.Url + pkg.Name + "/-/" + tbName
	log = log.WithValues("package", pkg.Name, "version", pkg.Version)
	log.Debug("identified")

	// get credentials and TODO cache it
	cred := npmCredentials.GetCredentials(ctx.GetContext(), b.spec.Url, pkg.Name)
	if cred == nil {
		return nil, fmt.Errorf("No credentials found for %s. Couldn't upload '%s'.", b.spec.Url, pkg.Name)
	}
	log.Debug("found credentials")

	// check if token exists, if not login and retrieve token
	token := cred[npmCredentials.ATTR_TOKEN]
	if token == "" {
		// use user+pass+mail from credentials to login and retrieve bearer token
		username := cred[npmCredentials.ATTR_USERNAME]
		password := cred[npmCredentials.ATTR_PASSWORD]
		email := cred[npmCredentials.ATTR_EMAIL]
		if username == "" || password == "" || email == "" {
			return nil, fmt.Errorf("No credentials for %s are invalid. Username, password or email missing! Couldn't upload '%s'.", b.spec.Url, pkg.Name)
		}
		log = log.WithValues("user", username, "repo", b.spec.Url)
		log.Debug("login")

		// TODO: check different kinds of .npmrc content
		token, err = login(b.spec.Url, username, password, email)
		if err != nil {
			return nil, err
		}
	} else {
		log.Debug("token found, skipping login")
	}

	// check if package exists
	exists, err := packageExists(b.spec.Url, *pkg, token)
	if err != nil {
		return nil, err
	}
	if exists {
		log.Debug("package+version already exists, skipping upload")
		return npm.New(b.spec.Url, pkg.Name, pkg.Version), nil
	}

	// prepare body for upload
	body := Body{
		ID:          pkg.Name,
		Name:        pkg.Name,
		Description: pkg.Description,
	}
	body.Versions = map[string]*Package{
		pkg.Version: pkg,
	}
	body.DistTags.Latest = pkg.Version
	body.Readme = pkg.Readme
	body.Attachments = map[string]*Attachment{
		tbName: NewAttachment(data),
	}
	marshal, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// prepare PUT request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, b.spec.Url+"/"+url.PathEscape(pkg.Name), bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("content-type", "application/json")

	// send PUT request - upload tgz
	client := http.Client{}
	log.Debug("uploading")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("http (%d) - failed to upload package: %s", resp.StatusCode, string(all))
	}
	log.Debug("successfully uploaded")
	return npm.New(b.spec.Url, pkg.Name, pkg.Version), nil
}

// Check if package already exists in npm registry. If it does, checks if it's the same.
func packageExists(repoUrl string, pkg Package, token string) (bool, error) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, repoUrl+"/"+url.PathEscape(pkg.Name)+"/"+url.PathEscape(pkg.Version), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		// artifact doesn't exist, it's safe to upload
		return false, nil
	}

	// artifact exists, let's check if it's the same
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("http (%d) - %s", resp.StatusCode, string(all))
	}
	var data map[string]interface{}
	err = json.Unmarshal(all, &data)
	if err != nil {
		return false, err
	}
	dist := data["dist"].(map[string]interface{})
	if pkg.Dist.Integrity == dist["integrity"] {
		// sha-512 sum is the same, we can skip the upload
		return true, nil
	}
	if pkg.Dist.Shasum == dist["shasum"] {
		// sha-1 sum is the same, we can skip the upload
		return true, nil
	}

	return false, fmt.Errorf("artifact already exists but has different shasum or integrity")
}
