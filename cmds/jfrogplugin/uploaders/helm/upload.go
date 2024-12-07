package helm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
	"ocm.software/ocm/api/ocm/plugin/ppi"
)

func Upload(
	ctx context.Context,
	data io.Reader,
	client *http.Client,
	url *url.URL,
	creds credentials.Credentials,
) (_ ppi.AccessSpec, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var req *http.Request
	var res *http.Response
	if req, err = http.NewRequestWithContext(ctx, http.MethodPut, url.String(), data); err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for upload: %w", err)
	}
	SetHeadersFromCredentials(req, creds)

	if res, err = client.Do(req); err != nil {
		return nil, fmt.Errorf("failed to store blob in artifactory: %w", err)
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if invalid := 200 > res.StatusCode || res.StatusCode >= 300; invalid {
		var responseBytes []byte
		if responseBytes, err = io.ReadAll(res.Body); err != nil {
			var body string
			if len(responseBytes) > 0 {
				body = fmt.Sprintf(": %s", string(responseBytes))
			}
			return nil, fmt.Errorf("invalid response (status %v)%s", res.StatusCode, body)
		}
	}

	uploadResponse := &ArtifactoryUploadResponse{}
	if err = json.NewDecoder(res.Body).Decode(uploadResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return uploadResponse.ToHelmAccessSpec()
}

type ArtifactoryUploadResponse struct {
	Repo        string `json:"repo,omitempty"`
	Path        string `json:"path,omitempty"`
	Created     string `json:"created,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	DownloadUri string `json:"downloadUri,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Size        string `json:"size,omitempty"`
	Checksums   struct {
		Sha1   string `json:"sha1,omitempty"`
		Sha256 string `json:"sha256,omitempty"`
		Md5    string `json:"md5,omitempty"`
	} `json:"checksums,omitempty"`
	Uri string `json:"uri"`
}

func (r *ArtifactoryUploadResponse) URL() string {
	if r.DownloadUri != "" {
		return r.DownloadUri
	}
	return r.Uri
}

func (r *ArtifactoryUploadResponse) ToHelmAccessSpec() (ppi.AccessSpec, error) {
	u := r.URL()
	urlp, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	chart := path.Base(urlp.Path)
	chart = strings.TrimSuffix(chart, path.Ext(chart))

	// this is needed so that the chart version constructor for OCM is happy
	// OCM encodes helm charts with a ":"...
	if idx := strings.LastIndex(chart, "-"); idx > 0 {
		chart = chart[:idx] + ":" + chart[idx+1:]
	}

	urlp.Path = ""
	urlp = urlp.JoinPath("artifactory", "api", "helm", r.Repo)
	repo := urlp.String()

	return helm.New(chart, repo), nil
}
