package helm

import (
	"fmt"
	"io"
	"net/http"
	"path"

	"golang.org/x/net/context"
	"ocm.software/ocm/api/credentials"
)

func ReindexChart(ctx context.Context, client *http.Client, artifactoryURL string,
	repository string,
	creds credentials.Credentials,
) (err error) {
	reindexURL, err := convertToReindexURL(artifactoryURL, repository)
	if err != nil {
		return fmt.Errorf("failed to convert to reindex URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reindexURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create reindex request: %w", err)
	}
	SetHeadersFromCredentials(req, creds)
	req.Header = req.Header.Clone()

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reindex chart: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		responseBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body but server returned %v: %w", res.StatusCode, err)
		}
		var body string
		if len(responseBytes) > 0 {
			body = fmt.Sprintf(": %s", string(responseBytes))
		}
		return fmt.Errorf("invalid response (status %v) while reindexing at %q: %s", res.StatusCode, reindexURL, body)
	}

	return nil
}

// convertToReindexURL converts the base URL and repository to a reindex URL.
// see https://jfrog.com/help/r/jfrog-rest-apis/calculate-helm-chart-index for the reindex API
func convertToReindexURL(baseURL string, repository string) (string, error) {
	u, err := ParseURLAllowNoScheme(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	u.Path = path.Join(u.Path, "api", "helm", repository, "reindex")
	return u.String(), nil
}
