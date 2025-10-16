package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"testing"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
)

func TestUpload(t *testing.T) {
	const artifactory = "https://mocked.artifactory.localhost:9999"
	const chartName, chartVersion, repo = "chart", "1.0.0", "repo"
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method is not PUT as expected by JFrog")
		}
		if user, pass, ok := r.BasicAuth(); !ok {
			t.Fatalf("invalid basic auth: %s - %s", user, pass)
		}
		res := ArtifactoryUploadResponse{
			Repo:        repo,
			DownloadUri: fmt.Sprintf("%s/path/to/%s/%s-%s.tgz", artifactory, repo, chartName, chartVersion),
		}
		data, err := json.Marshal(res)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	t.Cleanup(func() {
		srv.Close()
	})
	client := srv.Client()

	url, err := neturl.Parse(srv.URL)
	if err != nil {
		t.Fatalf("unexpected test client URL: %v", err)
	}

	ctx := context.Background()

	data := strings.NewReader("testdata")

	accessSpec, err := Upload(ctx, data, client, url, credentials.DirectCredentials{
		credentials.ATTR_USERNAME: "foo",
		credentials.ATTR_PASSWORD: "bar",
	}, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if typ := accessSpec.GetType(); typ != helm.Type {
		t.Fatalf("unexpected type: %v", typ)
	}

	helmAccessSpec, ok := accessSpec.(*helm.AccessSpec)
	if !ok {
		t.Fatalf("unexpected cast failure to helm access spec")
	}

	if specChart := helmAccessSpec.GetChartName(); specChart != chartName {
		t.Fatalf("unexpected chart name: %v", specChart)
	}
	if specVersion := helmAccessSpec.GetVersion(); specVersion != chartVersion {
		t.Fatalf("unexpected chart version: %v", specVersion)
	}

	if helmAccessSpec.HelmRepository != fmt.Sprintf("%s/artifactory/api/helm/%s", artifactory, repo) {
		t.Fatalf("expected an injected helm api reference to artifactory")
	}

}
