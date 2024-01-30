package npmjs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/npm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/mime"
)

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{repospec}
}

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, artType, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	attr := b.spec
	if attr == nil {
		return nil, nil
	}

	mimeType := blob.MimeType()
	if mime.MIME_TGZ != mimeType && mime.MIME_TGZ_ALT != mimeType {
		return nil, nil
	}

	blobReader, err := blob.Reader()
	if err != nil {
		panic(err)
	}
	defer blobReader.Close()

	data, err := io.ReadAll(blobReader)

	// read package.json from tarball to get package name and version
	var pkg *Package
	pkg, err = prepare(data)
	if err != nil {
		panic(err)
	}

	// use user+pass+mail from credentials to login and retrieve bearer token
	cred := GetCredentials(ctx.GetContext(), b.spec.Url, pkg.Name)
	username := cred[ATTR_USERNAME]
	password := cred[ATTR_PASSWORD]
	email := cred[ATTR_EMAIL]
	if username == "" || password == "" || email == "" {
		return nil, fmt.Errorf("username, password or email missing")
	}
	token, err := login(b.spec.Url, username, password, email)
	if err != nil {
		panic(err)
	}

	// prepare body for upload
	body := Body{
		ID:          pkg.Name,
		Name:        pkg.Name,
		Description: pkg.Description,
	}
	body.DistTags.Latest = pkg.Version
	body.Versions = map[string]*Package{
		pkg.Version: pkg,
	}
	body.Readme = pkg.Readme
	tbName := pkg.Name + "-" + pkg.Version + ".tgz"
	body.Attachments = map[string]*Attachment{
		tbName: NewAttachment(data),
	}
	pkg.Dist.Tarball = b.spec.Url + pkg.Name + "/-/" + tbName
	marshal, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	// prepare request
	req, err := http.NewRequest(http.MethodPut, b.spec.Url+"/"+url.PathEscape(pkg.Name), bytes.NewReader(marshal))
	if err != nil {
		panic(err)
	}
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("content-type", "application/json")

	// send request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(all))
	}

	return npm.New(b.spec.Url, pkg.Name, pkg.Version), nil
}
