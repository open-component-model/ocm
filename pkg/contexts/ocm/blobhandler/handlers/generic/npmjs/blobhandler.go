package npmjs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/npm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
)

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{repospec}
}

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, artType, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	if b.spec == nil {
		return nil, nil
	}

	mimeType := blob.MimeType()
	if mime.MIME_TGZ != mimeType && mime.MIME_TGZ_ALT != mimeType {
		return nil, nil
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
	log := logging.Context().Logger()
	log.Debug("reading package.json from tarball")
	var pkg *Package
	pkg, err = prepare(data)
	if err != nil {
		return nil, err
	}
	tbName := pkg.Name + "-" + pkg.Version + ".tgz"
	pkg.Dist.Tarball = b.spec.Url + pkg.Name + "/-/" + tbName
	log.Debug("identified package: %s@%s", pkg.Name, pkg.Version)

	// use user+pass+mail from credentials to login and retrieve bearer token
	cred := GetCredentials(ctx.GetContext(), b.spec.Url, pkg.Name)
	username := cred[ATTR_USERNAME]
	password := cred[ATTR_PASSWORD]
	email := cred[ATTR_EMAIL]
	if username == "" || password == "" || email == "" {
		return nil, fmt.Errorf("username, password or email missing")
	}
	log.Debug("logging in as %s on %s", username, b.spec.Url)
	token, err := login(b.spec.Url, username, password, email)
	if err != nil {
		return nil, err
	}

	// check if package exists
	exists, err := packageExists(b.spec.Url, *pkg, token)
	if err != nil {
		return nil, err
	}
	if exists {
		log.Debug("package %s@%s already exists, skipping upload", pkg.Name, pkg.Version)
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
	log.Debug("uploading package %s@%s to %s", pkg.Name, pkg.Version, b.spec.Url)
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
		log.Error("http (%d) - failed to upload package: %s", resp.StatusCode, string(all))
	}
	log.Debug("package %s@%s uploaded to %s", pkg.Name, pkg.Version, b.spec.Url)
	return npm.New(b.spec.Url, pkg.Name, pkg.Version), nil
}
