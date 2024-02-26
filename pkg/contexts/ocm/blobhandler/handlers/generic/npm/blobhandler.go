// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
)

const BLOB_HANDLER_NAME = "ocm/npmPackage"

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
