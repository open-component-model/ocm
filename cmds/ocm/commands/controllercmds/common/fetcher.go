package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func getLatestVersion(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/latest", nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API call failed: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	type meta struct {
		Tag string `json:"tag_name"`
	}
	var m meta
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", fmt.Errorf("decoding GitHub API response failed: %w", err)
	}

	return m.Tag, err
}

// existingVersion calls the GitHub API to confirm the given version does exist.
func existingVersion(ctx context.Context, url, version string) (bool, error) {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	ghURL := fmt.Sprintf(url+"/tags/%s", version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ghURL, nil)
	if err != nil {
		return false, fmt.Errorf("GitHub API call failed: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("GitHub API returned an unexpected status code (%d)", res.StatusCode)
	}
}

func fetch(ctx context.Context, url, version, dir, filename string) error {
	ghURL := fmt.Sprintf("%s/latest/download/%s", url, filename)
	if strings.HasPrefix(version, "v") {
		ghURL = fmt.Sprintf("%s/download/%s/%s", url, version, filename)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ghURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %s, error: %w", ghURL, err)
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to download manifests.tar.gz from %s, error: %w", ghURL, err)
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s from %s, status: %s", filename, ghURL, resp.Status)
	}

	wf, err := os.OpenFile(filepath.Join(dir, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o777)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}

	if _, err := io.Copy(wf, resp.Body); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	return nil
}
