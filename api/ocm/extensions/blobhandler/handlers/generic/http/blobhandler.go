package maven

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"text/template"

	"github.com/google/go-containerregistry/pkg/v1/remote"
	mlog "github.com/mandelsoft/logging"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/generic/http/identity"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/logging"
)

const REALM = "http"

const BlobHandlerName = "ocm/" + "http"

type artifactHandler struct {
	spec *Config
}

func NewArtifactHandler(repospec *Config) cpi.BlobHandler {
	return &artifactHandler{spec: repospec}
}

// blob => http.Request => http.Response => cpi.AccessSpec (HelmAccess, WgetAccess, etc.)
func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, resourceType string, hint string, _ cpi.AccessSpec, ctx cpi.StorageContext) (_ cpi.AccessSpec, rerr error) {
	remote, err := b.URL(ctx, blob)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote for blob: %w", err)
	}
	data, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}
	defer func() {
		rerr = errors.Join(rerr, data.Close())
	}()

	rawURL := remote.String()

	req, err := http.NewRequestWithContext(context.TODO(), b.spec.Method, rawURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	creds := identity.GetCredentials(ctx.GetContext(), rawURL)
	user, pass := creds[identity.ATTR_USERNAME], creds[identity.ATTR_PASSWORD]
	if user != "" && pass != "" {
		req.SetBasicAuth(creds["username"], creds["password"])
	}

	client := Client(b.spec)
	client.Transport = logging.NewRoundTripper(client.Transport, logging.DynamicLogger(ctx, REALM,
		mlog.NewAttribute(logging.ATTR_HOST, remote.Host),
		mlog.NewAttribute(logging.ATTR_PATH, remote.Path),
		mlog.NewAttribute(logging.ATTR_USER, user),
	))

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to store blob via request: %w", err)
	}

	//TODO: replace with registry based mapping so we dont have to bind the api here
	switch resourceType {
	case resourcetypes.HELM_CHART:
		return helm.New("demoapp", "https://int.repositories.cloud.sap/artifactory/api/helm/ocm-helm-test"), nil
	}

	return nil, errors.New("not implemented")
}

func (h *artifactHandler) URL(ctx cpi.StorageContext, blob cpi.BlobAccess) (*url.URL, error) {
	tpl, err := template.New("").Parse(h.spec.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url template: %w", err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]string{
		"mimeType":  blob.MimeType(),
		"size":      fmt.Sprintf("%d", blob.Size()),
		"digest":    string(blob.Digest()),
		"component": ctx.TargetComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute url template: %w", err)
	}
	return utils.ParseURL(buf.String())
}

func Client(_ *Config) *http.Client {
	return &http.Client{
		Transport: remote.DefaultTransport,
	}
}
