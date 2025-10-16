package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	godigest "github.com/opencontainers/go-digest"

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
	digest string,
) (_ ppi.AccessSpec, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var req *http.Request
	var res *http.Response
	if req, err = http.NewRequestWithContext(ctx, http.MethodPut, url.String(), data); err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for upload: %w", err)
	}

	// if there is no digest information, we skip the digest headers.
	// note that this will cause insecure uploads and should be avoided where possible
	if digest != "" {
		parsedDigest, err := godigest.Parse(digest)
		if err != nil {
			return nil, fmt.Errorf("failed to parse digest: %w", err)
		}

		// see https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact-by-checksum for the checksum headers
		switch parsedDigest.Algorithm() {
		case godigest.SHA256:
			req.Header.Set("X-Checksum-Sha256", parsedDigest.Encoded())
		default:
			return nil, fmt.Errorf("unsupported digest algorithm, must be %s to allow upload to jfrog: %s", godigest.SHA256, parsedDigest.Algorithm())
		}
		req.Header.Set("X-Checksum-Deploy", "false")
	}

	req.Header.Set("Accept-Encoding", "gzip")

	SetHeadersFromCredentials(req, creds)

	if res, err = client.Do(req); err != nil {
		return nil, fmt.Errorf("failed to store blob in artifactory: %w", err)
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if res.StatusCode != http.StatusCreated {
		responseBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body but server returned %v: %w", res.StatusCode, err)
		}
		var body string
		if len(responseBytes) > 0 {
			body = fmt.Sprintf(": %s", string(responseBytes))
		}
		return nil, fmt.Errorf("invalid response (status %v)%s", res.StatusCode, body)
	}

	var buf bytes.Buffer
	body := io.TeeReader(res.Body, &buf)
	uploadResponse := &ArtifactoryUploadResponse{}
	if err = json.NewDecoder(body).Decode(uploadResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response (original %q): %w", buf.String(), err)
	}

	return uploadResponse.ToHelmAccessSpec()
}

type ArtifactoryUploadResponse struct {
	Repo        string                     `json:"repo,omitempty"`
	Path        string                     `json:"path,omitempty"`
	Created     string                     `json:"created,omitempty"`
	CreatedBy   string                     `json:"createdBy,omitempty"`
	DownloadUri string                     `json:"downloadUri,omitempty"`
	MimeType    string                     `json:"mimeType,omitempty"`
	Size        string                     `json:"size,omitempty"`
	Checksums   ArtifactoryUploadChecksums `json:"checksums,omitempty"`
	Uri         string                     `json:"uri"`
}

type ArtifactoryUploadChecksums struct {
	Sha1   string `json:"sha1,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
	Md5    string `json:"md5,omitempty"`
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
